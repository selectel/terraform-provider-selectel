package mks

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/stretchr/testify/require"
)

const testClusterID = "5340b834-1a2b-4c2f-9d0a-89d24d99c0b8"

type fakeAPI struct {
	t  *testing.T
	mu sync.Mutex

	server *httptest.Server

	clusters           map[string]mksclient.ClusterDetailed
	kubeVersions       []mksclient.KubeVersionInfo
	failures           map[string]int
	failNext           map[string]fakeFailure
	failureBodies      map[string]fakeBody
	tasks              []*fakeTask
	failedTaskTypes    map[string]bool
	stuckTaskTypes     map[string]bool
	settles            map[string]fakeSettle
	x509CACertificates map[string]string
	queries            map[string]string
	bodies             map[string][]byte
	calls              map[string]int
}

type fakeBody struct {
	contentType string
	body        string
}

// fakeFailure is a transient route failure: status applies to the next left
// requests.
type fakeFailure struct {
	status int
	left   int
}

type fakeSettle struct {
	gets   int
	status string
}

type fakeTask struct {
	task   mksclient.Task
	stuck  bool
	effect func()
}

const (
	keystoneProject = "35b965e88e5948488eed5e509833033c"
	apiProject      = "35b965e8-8e59-4848-8eed-5e509833033c"
)

const (
	routeCluster       = "GET /v2/clusters/{cluster_id}"
	routeKubeVersions  = "GET /v2/kubeversions"
	routeCreateCluster = "POST /v2/clusters"
	routePatchCluster  = "PATCH /v2/clusters/{cluster_id}"
	routeDeleteCluster = "DELETE /v2/clusters/{cluster_id}"
	routeUpgradePatch  = "POST /v2/clusters/{cluster_id}/upgrade-patch-version"
	routeUpgradeMinor  = "POST /v2/clusters/{cluster_id}/upgrade-minor-version"
	routeTasks         = "GET /v2/clusters/{cluster_id}/tasks"
	routeTask          = "GET /v2/clusters/{cluster_id}/tasks/{task_id}"
)

var shortPollInterval sync.Once

func newFakeAPI(t *testing.T) *fakeAPI {
	t.Helper()

	f := &fakeAPI{
		t: t,

		clusters:      map[string]mksclient.ClusterDetailed{},
		failures:      map[string]int{},
		failNext:      map[string]fakeFailure{},
		failureBodies: map[string]fakeBody{},

		failedTaskTypes:    map[string]bool{},
		stuckTaskTypes:     map[string]bool{},
		settles:            map[string]fakeSettle{},
		x509CACertificates: map[string]string{},
		queries:            map[string]string{},
		bodies:             map[string][]byte{},
		calls:              map[string]int{},
	}

	mux := http.NewServeMux()
	f.handle(mux, routeCluster, f.getCluster)
	f.handle(mux, routeKubeVersions, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, mksclient.KubeVersionsList{KubeVersions: &f.kubeVersions})
	})
	f.handle(mux, routeCreateCluster, f.createCluster)
	f.handle(mux, routePatchCluster, f.patchCluster)
	f.handle(mux, routeDeleteCluster, f.deleteCluster)
	f.handle(mux, routeUpgradePatch, f.upgradeCluster(false))
	f.handle(mux, routeUpgradeMinor, f.upgradeCluster(true))
	f.handle(mux, routeTasks, f.listTasks)
	f.handle(mux, routeTask, f.getTask)
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)

	shortPollInterval.Do(func() { pollInterval = time.Millisecond })

	return f
}

func (f *fakeAPI) client() *mksv2.ServiceClient {
	t := f.t
	t.Helper()
	client, err := mksv2.NewMKSClientV2("fake-token", f.server.URL)
	require.NoError(t, err)

	return client
}

func (f *fakeAPI) handle(mux *http.ServeMux, pattern string, handler http.HandlerFunc) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()

		f.calls[pattern]++
		f.queries[pattern] = r.URL.RawQuery
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			f.bodies[pattern] = body
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		if ff, ok := f.failNext[pattern]; ok && ff.left > 0 {
			f.failNext[pattern] = fakeFailure{status: ff.status, left: ff.left - 1}
			objectType, id := fakeObject(r)
			writeAPIError(f.t, w, ff.status, objectType, id)

			return
		}
		status, ok := f.failures[pattern]
		raw, hasBody := f.failureBodies[pattern]
		if ok && hasBody {
			w.Header().Set("Content-Type", raw.contentType)
			w.WriteHeader(status)
			_, _ = io.WriteString(w, raw.body)

			return
		}
		if ok {
			objectType, id := fakeObject(r)
			writeAPIError(f.t, w, status, objectType, id)

			return
		}

		handler(w, r)
	})
}

func (f *fakeAPI) getCluster(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("cluster_id")
	status, settled := f.settle(id)
	if settled {
		f.updateClusterLocked(id, status)
	}
	c, ok := f.clusters[id]
	if !ok {
		writeAPIError(f.t, w, http.StatusNotFound, objectCluster, id)

		return
	}

	writeJSON(w, http.StatusOK, mksclient.ClusterResp{Cluster: &c})
}

func (f *fakeAPI) createCluster(w http.ResponseWriter, r *http.Request) {
	var body mksclient.ClusterCreateBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.Cluster == nil {
		writeAPIError(f.t, w, http.StatusBadRequest, objectCluster, "")

		return
	}
	opts := body.Cluster

	clusterType := mksclient.HIGHAVAILABILITY
	if opts.ClusterType != nil {
		clusterType = mksclient.ClusterDetailedClusterType(*opts.ClusterType)
	}
	cniType := mksclient.ClusterDetailedCniType(mksclient.ClusterCniTypeCALICO)
	if opts.CniType != nil {
		cniType = mksclient.ClusterDetailedCniType(*opts.CniType)
	}
	c := mksclient.ClusterDetailed{
		Id:                            testClusterID,
		Name:                          strings.ToLower(opts.Name),
		Pool:                          opts.Pool,
		ProjectId:                     fakeAPIProjectID(keystoneProject),
		KubeVersion:                   opts.KubeVersion,
		ClusterType:                   clusterType,
		Basic:                         clusterType == mksclient.BASIC,
		NetworkType:                   mksclient.ClusterDetailedNetworkType(opts.NetworkType),
		NetworkId:                     cmp.Or(opts.NetworkId, "fake-network"),
		SubnetId:                      cmp.Or(opts.SubnetId, "fake-subnet"),
		PrivateKubeApi:                opts.PrivateKubeApi,
		EnableAutorepair:              valueOrT(opts.EnableAutorepair, true),
		EnablePatchVersionAutoUpgrade: valueOrT(opts.EnablePatchVersionAutoUpgrade, clusterType != mksclient.BASIC),
		CniType:                       cniType,
		KubeApiIp:                     "192.0.2.10",
		Status:                        "PENDING_CREATE",
	}
	f.setWindow(&c, cmp.Or(opts.MaintenanceWindowStart, "03:00:00"))
	if cniType == mksclient.ClusterDetailedCniType(mksclient.ClusterCniTypeCILIUM) {
		cilium := valueOrZero(opts.CniCiliumSettings)
		c.CniCiliumSettings = &mksclient.CNICiliumSettings{
			EnvoyDaemonset: new(valueOrT(cilium.EnvoyDaemonset, true)),
			HubbleRelay:    new(valueOrT(cilium.HubbleRelay, true)),
		}
	}
	if opts.KubernetesOptions != nil {
		c.KubernetesOptions = *opts.KubernetesOptions
		c.KubernetesOptions.Oidc, err = normaliseFakeOIDC(c.KubernetesOptions.Oidc)
		if err != nil {
			writeFakeBadRequest(f.t, w, err)

			return
		}
		c.KubernetesOptions.X509CaCertificates = ""
	}
	f.clusters[c.Id] = c

	f.addTask(c.Id, "CREATE_CLUSTER", f.activateCluster(c.Id))
	writeJSON(w, http.StatusCreated, mksclient.ClusterResp{Cluster: &c})
}

func (f *fakeAPI) patchCluster(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("cluster_id")
	c, ok := f.clusters[id]
	if !ok {
		writeAPIError(f.t, w, http.StatusNotFound, objectCluster, id)

		return
	}
	var body mksclient.ClusterUpdateBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.Cluster == nil {
		writeAPIError(f.t, w, http.StatusBadRequest, objectCluster, id)

		return
	}
	opts := body.Cluster

	if opts.MaintenanceWindowStart != nil {
		f.setWindow(&c, *opts.MaintenanceWindowStart)
	}
	c.EnableAutorepair = valueOrT(opts.EnableAutorepair, c.EnableAutorepair)
	c.EnablePatchVersionAutoUpgrade = valueOrT(opts.EnablePatchVersionAutoUpgrade, c.EnablePatchVersionAutoUpgrade)
	if opts.CniCiliumSettings != nil {
		cilium := valueOrZero(c.CniCiliumSettings)
		cilium.EnvoyDaemonset = cmp.Or(opts.CniCiliumSettings.EnvoyDaemonset, cilium.EnvoyDaemonset)
		cilium.HubbleRelay = cmp.Or(opts.CniCiliumSettings.HubbleRelay, cilium.HubbleRelay)
		c.CniCiliumSettings = &cilium
		c.Status = "PENDING_UPGRADE_CLUSTER_CONFIG"
		f.addTask(id, "UPGRADE_CLUSTER_CONFIG", f.activateCluster(id))
	}
	if opts.KubernetesOptions != nil {
		changed, err := f.patchKubeOptions(id, &c, *opts.KubernetesOptions)
		if err != nil {
			writeFakeBadRequest(f.t, w, err)

			return
		}
		if changed {
			c.Status = "PENDING_UPGRADE_CLUSTER_CONFIG"
			f.addTask(id, "UPGRADE_MASTERS_CONFIG", f.activateCluster(id))
		}
	}
	f.clusters[id] = c

	writeJSON(w, http.StatusOK, mksclient.ClusterResp{Cluster: &c})
}

func (f *fakeAPI) patchKubeOptions(id string, c *mksclient.ClusterDetailed, requested mksclient.KubernetesOptions) (bool, error) {
	current := &c.KubernetesOptions
	var err error
	requested.Oidc, err = normaliseFakeOIDC(requested.Oidc)
	if err != nil {
		return false, err
	}
	changed := false

	if !slices.Equal(current.FeatureGates, requested.FeatureGates) {
		current.FeatureGates = requested.FeatureGates
		changed = true
	}
	if !slices.Equal(current.AdmissionControllers, requested.AdmissionControllers) {
		current.AdmissionControllers = requested.AdmissionControllers
		changed = true
	}
	if current.AuditLogs.Enabled != requested.AuditLogs.Enabled {
		current.AuditLogs = requested.AuditLogs
		changed = true
	}
	oidcChanged := current.Oidc
	oidcChanged.ProviderName = requested.Oidc.ProviderName
	if oidcChanged != requested.Oidc {
		current.Oidc = requested.Oidc
		changed = true
	}
	if requested.X509CaCertificates != "" {
		f.x509CACertificates[id] = requested.X509CaCertificates
		changed = true
	}

	return changed, nil
}

var errFakeOIDCDisabled = errors.New("oidc parameters cannot be configured when it is disabled")

func normaliseFakeOIDC(oidc mksclient.OIDC) (mksclient.OIDC, error) {
	if !oidc.Enabled {
		if oidc != (mksclient.OIDC{}) {
			return mksclient.OIDC{}, errFakeOIDCDisabled
		}

		return mksclient.OIDC{}, nil
	}
	oidc.UsernameClaim = cmp.Or(oidc.UsernameClaim, "sub")
	oidc.GroupsClaim = cmp.Or(oidc.GroupsClaim, "groups")
	oidc.CaCerts = strings.TrimSpace(oidc.CaCerts)

	return oidc, nil
}

func (f *fakeAPI) activateCluster(id string) func() {
	return func() {
		c, ok := f.clusters[id]
		if ok {
			c.Status = "ACTIVE"
			f.clusters[id] = c
		}
	}
}

func (f *fakeAPI) deleteCluster(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("cluster_id")
	c, ok := f.clusters[id]
	if !ok {
		writeAPIError(f.t, w, http.StatusNotFound, objectCluster, id)

		return
	}
	if !slices.Contains(clusterDeleteStatuses, string(c.Status)) {
		var body mksclient.GenericError
		body.Error.Message = "cluster or one of nodegroups has invalid status to perform this operation"
		writeJSON(w, http.StatusConflict, body)

		return
	}

	c.Status = "PENDING_DELETE"
	f.clusters[id] = c
	f.addTask(id, "DELETE_CLUSTER", func() { delete(f.clusters, id) })
	w.WriteHeader(http.StatusNoContent)
}

func (f *fakeAPI) upgradeCluster(minor bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("cluster_id")
		c, ok := f.clusters[id]
		if !ok {
			writeAPIError(f.t, w, http.StatusNotFound, objectCluster, id)

			return
		}

		target, taskType := kubeVersionTrimToMinor, "UPGRADE_PATCH_VERSION"
		if minor {
			target, taskType = kubeVersionTrimToMinorIncremented, "UPGRADE_MINOR_VERSION"
		}
		targetMinor, err := target(c.KubeVersion)
		if err != nil {
			writeAPIError(f.t, w, http.StatusBadRequest, objectCluster, id)

			return
		}
		latest := map[string]string{}
		for _, v := range f.kubeVersions {
			minor, err := kubeVersionTrimToMinor(valueOrZero(v.Version))
			if err != nil {
				continue
			}
			// The API upgrades to the latest patch of the target minor.
			current, ok := latest[minor]
			if !ok {
				latest[minor] = valueOrZero(v.Version)
				continue
			}
			newer, err := compareTwoKubeVersionsByPatch(valueOrZero(v.Version), current)
			if err == nil && newer == valueOrZero(v.Version) {
				latest[minor] = newer
			}
		}
		if latest[targetMinor] == "" {
			writeAPIError(f.t, w, http.StatusBadRequest, objectCluster, id)

			return
		}

		f.addTask(id, taskType, func() {
			c := f.clusters[id]
			c.KubeVersion = latest[targetMinor]
			f.clusters[id] = c
		})
		writeJSON(w, http.StatusOK, mksclient.ClusterResp{Cluster: &c})
	}
}

func (f *fakeAPI) listTasks(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("cluster_id")
	_, ok := f.clusters[id]
	if !ok {
		writeAPIError(f.t, w, http.StatusNotFound, objectCluster, id)

		return
	}

	var tasks []mksclient.Task
	for _, t := range slices.Backward(f.tasks) {
		if t.task.ClusterId == id {
			tasks = append(tasks, t.task)
		}
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	tasks = tasks[min(offset, len(tasks)):]
	if limit > 0 {
		tasks = tasks[:min(limit, len(tasks))]
	}

	writeJSON(w, http.StatusOK, mksclient.TaskList{Count: int64(len(tasks)), Tasks: tasks})
}

func (f *fakeAPI) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("cluster_id")
	taskID := r.PathValue("task_id")
	i := slices.IndexFunc(f.tasks, func(t *fakeTask) bool { return t.task.Id == taskID && t.task.ClusterId == id })
	_, clusterExists := f.clusters[id]
	if !clusterExists {
		writeAPIError(f.t, w, http.StatusNotFound, objectCluster, id)

		return
	}
	if i < 0 {
		writeAPIError(f.t, w, http.StatusNotFound, objectTask, taskID)

		return
	}

	t := f.tasks[i]
	running := t.task.Status == mksclient.INQUEUE || t.task.Status == mksclient.INPROGRESS
	switch {
	case !running || t.stuck || f.stuckTaskTypes[t.task.Type]:
	case f.failedTaskTypes[t.task.Type]:
		t.task.Status = mksclient.ERROR
		t.task.ErrorDetails = &mksclient.TaskErrorDetails{
			Code: 42, Name: "fake failure", Details: new("fake details of " + t.task.Type),
		}
		f.updateClusterLocked(id, "ERROR")
	default:
		t.task.Status = mksclient.DONE
		if t.effect != nil {
			t.effect()
		}
	}

	writeJSON(w, http.StatusOK, mksclient.TaskResp{Task: t.task})
}

func (f *fakeAPI) addTask(clusterID, taskType string, effect func()) {
	f.tasks = append(f.tasks, &fakeTask{
		task: mksclient.Task{
			Id:        fmt.Sprintf("task-%d", len(f.tasks)+1),
			ClusterId: clusterID,
			Type:      taskType,
			Status:    mksclient.INPROGRESS,
		},
		effect: effect,
	})
}

func (f *fakeAPI) seedCluster(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.clusters[id] = mksclient.ClusterDetailed{
		Id:        id,
		ProjectId: fakeAPIProjectID(keystoneProject),
		Pool:      "ru-1",
		Status:    "ACTIVE",
	}
}

func (f *fakeAPI) seedTask(clusterID, nodegroupID, taskType string, stuck bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.addTask(clusterID, taskType, nil)
	t := f.tasks[len(f.tasks)-1]
	t.stuck = stuck
	if nodegroupID != "" {
		t.task.NodegroupId = &nodegroupID
	}
}

func (f *fakeAPI) stickTasks(taskType string, stuck bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.stuckTaskTypes[taskType] = stuck
}

func (f *fakeAPI) failTasks(taskType string, failed bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.failedTaskTypes[taskType] = failed
}

func (f *fakeAPI) lastBody(route string) map[string]any {
	f.t.Helper()

	f.mu.Lock()
	defer f.mu.Unlock()

	var body map[string]any
	err := json.Unmarshal(f.bodies[route], &body)
	require.NoError(f.t, err, "decode the last %s body %q", route, f.bodies[route])

	return body
}

func (f *fakeAPI) callCount(route string) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls[route]
}

func (f *fakeAPI) clusterCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.clusters)
}

func (f *fakeAPI) updateCluster(id string, update func(c *mksclient.ClusterDetailed)) {
	f.mu.Lock()
	defer f.mu.Unlock()

	c := f.clusters[id]
	update(&c)
	f.clusters[id] = c
}

func (f *fakeAPI) removeCluster(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.clusters, id)
}

func (f *fakeAPI) clusterProject(id string) string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.clusters[id].ProjectId
}

func (f *fakeAPI) settleAfter(id string, gets int, status string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.settles[id] = fakeSettle{gets: gets, status: status}
}

func (f *fakeAPI) settle(id string) (string, bool) {
	s, ok := f.settles[id]
	if !ok {
		return "", false
	}
	s.gets--
	if s.gets > 0 {
		f.settles[id] = s

		return "", false
	}
	delete(f.settles, id)

	return s.status, true
}

func (f *fakeAPI) updateClusterLocked(id, status string) {
	c, ok := f.clusters[id]
	if !ok {
		return
	}
	if status == "" {
		delete(f.clusters, id)

		return
	}
	c.Status = mksclient.ClusterDetailedStatus(status)
	f.clusters[id] = c
}

func (f *fakeAPI) setWindow(c *mksclient.ClusterDetailed, start string) {
	c.MaintenanceWindowStart = start
	parsed, err := time.Parse(time.TimeOnly, start)
	if err != nil {
		c.MaintenanceWindowEnd = ""

		return
	}
	c.MaintenanceWindowEnd = parsed.Add(3 * time.Hour).Format(time.TimeOnly)
}

func (f *fakeAPI) seedKubeVersions(versions ...mksclient.KubeVersionInfo) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.kubeVersions = versions
}

func (f *fakeAPI) fail(route string, status int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.failureBodies, route)
	if status == 0 {
		delete(f.failures, route)

		return
	}
	f.failures[route] = status
}

func (f *fakeAPI) failFor(route string, status, requests int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.failNext[route] = fakeFailure{status: status, left: requests}
}

func (f *fakeAPI) failWithBody(route string, status int, contentType, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.failures[route] = status
	f.failureBodies[route] = fakeBody{contentType: contentType, body: body}
}

func (f *fakeAPI) lastQuery(route string) string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.queries[route]
}

func (f *fakeAPI) storedX509(id string) string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.x509CACertificates[id]
}

func fakeAPIProjectID(projectID string) string {
	if len(projectID) != 32 {
		return projectID
	}

	return projectID[:8] + "-" + projectID[8:12] + "-" + projectID[12:16] + "-" + projectID[16:20] + "-" + projectID[20:]
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

const (
	objectTask = "Task"
)

func fakeObject(r *http.Request) (string, string) {
	id := r.PathValue("task_id")
	if id != "" {
		return objectTask, id
	}

	return objectCluster, r.PathValue("cluster_id")
}

func writeAPIError(t *testing.T, w http.ResponseWriter, status int, objectType, id string) {
	t.Helper()
	if status == http.StatusNotFound {
		var body mksclient.GenericNotFoundError
		body.Error.Id = id
		body.Error.Message = objectType + " not found"
		writeJSON(w, status, body)

		return
	}

	var body mksclient.GenericError
	body.Error.Message = fmt.Sprintf("fake error %d", status)
	writeJSON(w, status, body)
}

func writeFakeBadRequest(t *testing.T, w http.ResponseWriter, err error) {
	t.Helper()
	var body mksclient.GenericError
	body.Error.Message = err.Error()
	writeJSON(w, http.StatusBadRequest, body)
}

func valueOrT[T any](v *T, fallback T) T {
	if v == nil {
		return fallback
	}

	return *v
}
