package filestorage

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const networkShareID = "44444444-4444-4444-8444-444444444444"

// fakeAPI serves one share of the File Storage API with its service
// share network and mount paths, and records every request.
type fakeAPI struct {
	t  *testing.T
	mu sync.Mutex

	share   map[string]any // nil when the share does not exist
	subnets int            // subnets of the share network; 0 when the network is gone
	paths   []map[string]any

	// statuses are returned by successive GETs of the share; the last repeats.
	// "rejected" reads as available at the old size, as after a rejected extension.
	statuses []string
	// newSize is the requested size, which the share gets once a GET shows a
	// status other than rejected.
	newSize int
	// fail answers the matching "METHOD path" with the status and body once set.
	fail map[string]answer
	// failOnce answers the matching "METHOD path" with its answers in turn and
	// then serves the request; a zero answer serves that request too.
	failOnce map[string][]answer
	// lostCreate creates the share but answers the create request with this failure.
	lostCreate *answer
	// rejectCreate answers the create request without creating the share.
	rejectCreate *answer
	// hiddenLists is the number of share lists that do not show the share yet.
	hiddenLists int
	// others are more shares of the project, listed after share.
	others []map[string]any
	// messages are the user messages of the project, newest first.
	messages []map[string]any

	requests []string
	bodies   map[string]string
	queries  map[string]url.Values
}

type answer struct {
	status int
	body   string
}

// unavailable503 is a transient failure that the waits retry.
var unavailable503 = answer{http.StatusServiceUnavailable, `{"message": "Service Unavailable"}`}

const apiFault400 = `{"badRequest": {"code": 400, "message": "IP 10.0.0.5 is already used."}}`

func newFakeAPI(t *testing.T) *fakeAPI {
	t.Helper()

	return &fakeAPI{
		t: t,
		share: map[string]any{
			"id": shareID, "status": "available", "name": "data", "description": "",
			"share_proto": "NFS", "size": 50, "share_type": "type-uuid", "share_type_name": "basic.ru-9a",
			"availability_zone": "ru-9a", "share_network_id": networkShareID,
			"metadata": map[string]any{deleteNetworkKey: "enabled", "neutron_subnet_id": subnetID},
		},
		subnets:  1,
		paths:    []map[string]any{{"id": "el-1", "path": "10.0.0.5:/shares/share-1", "preferred": true}},
		fail:     map[string]answer{},
		failOnce: map[string][]answer{},
		bodies:   map[string]string{},
		queries:  map[string]url.Values{},
	}
}

func (f *fakeAPI) client() *sdk.Client {
	server := httptest.NewServer(http.HandlerFunc(f.serve))
	f.t.Cleanup(server.Close)
	client, err := sdk.NewClient(sdk.Config{Endpoint: server.URL, Token: "token"})
	require.NoError(f.t, err)

	return client
}

func (f *fakeAPI) resource() *shareResource { return testShareResource(f.client()) }

func (f *fakeAPI) count(request string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, r := range f.requests {
		if r == request {
			n++
		}
	}

	return n
}

func (f *fakeAPI) writes() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var writes []string
	for _, r := range f.requests {
		if !strings.HasPrefix(r, http.MethodGet) {
			writes = append(writes, r)
		}
	}

	return writes
}

func (f *fakeAPI) serve(w http.ResponseWriter, req *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	assert.Equal(f.t, "token", req.Header.Get("X-Auth-Token"))
	key := req.Method + " " + req.URL.EscapedPath()
	body, _ := io.ReadAll(req.Body)
	f.requests = append(f.requests, key)
	f.bodies[key] = string(body)
	f.queries[key] = req.URL.Query()
	w.Header().Set("Content-Type", "application/json")
	if failure, ok := f.fail[key]; ok {
		w.WriteHeader(failure.status)
		_, _ = io.WriteString(w, failure.body)

		return
	}
	if failures := f.failOnce[key]; len(failures) > 0 {
		f.failOnce[key] = failures[1:]
		if failure := failures[0]; failure.status != 0 {
			w.WriteHeader(failure.status)
			_, _ = io.WriteString(w, failure.body)

			return
		}
	}
	sharePath := "/shares/" + shareID
	switch {
	case key == "POST /shares":
		f.create(w, body)
	case key == "GET /shares/detail":
		f.list(w, req)
	case key == "GET /messages":
		f.reply(w, http.StatusOK, map[string]any{"messages": f.messages})
	case key == "GET "+sharePath:
		if f.share == nil {
			f.notFound(w)
			return
		}
		if len(f.statuses) > 0 {
			f.share["status"] = f.statuses[0]
			if len(f.statuses) > 1 {
				f.statuses = f.statuses[1:]
			}
		}
		if f.share["status"] == "rejected" {
			f.share["status"] = "available"
		} else if f.newSize != 0 {
			f.share["size"] = f.newSize
		}
		if f.share["status"] == "deleted" {
			f.share = nil
			f.notFound(w)

			return
		}
		f.reply(w, http.StatusOK, map[string]any{"share": f.share})
	case key == "PUT "+sharePath:
		var update struct {
			Share map[string]string `json:"share"`
		}
		require.NoError(f.t, json.Unmarshal(body, &update))
		if name, ok := update.Share["display_name"]; ok {
			f.share["name"] = name
		}
		if description, ok := update.Share["display_description"]; ok {
			f.share["description"] = description
		}
		f.reply(w, http.StatusOK, map[string]any{"share": f.share})
	case key == "POST "+sharePath+"/action":
		var action struct {
			Extend struct {
				NewSize int `json:"new_size"`
			} `json:"extend"`
		}
		require.NoError(f.t, json.Unmarshal(body, &action))
		f.newSize = action.Extend.NewSize
		w.WriteHeader(http.StatusAccepted)
	case key == "DELETE "+sharePath:
		f.share["status"] = "deleting"
		w.WriteHeader(http.StatusAccepted)
	case key == "POST "+sharePath+"/metadata":
		var update struct {
			Metadata map[string]string `json:"metadata"`
		}
		require.NoError(f.t, json.Unmarshal(body, &update))
		metadata := f.share["metadata"].(map[string]any)
		for k, v := range update.Metadata {
			metadata[k] = v
		}
		f.reply(w, http.StatusOK, map[string]any{"metadata": metadata})
	case strings.HasPrefix(key, "DELETE "+sharePath+"/metadata/"):
		delete(f.share["metadata"].(map[string]any), strings.TrimPrefix(key, "DELETE "+sharePath+"/metadata/"))
		w.WriteHeader(http.StatusOK)
	case key == "GET "+sharePath+"/export_locations":
		f.reply(w, http.StatusOK, map[string]any{"export_locations": f.paths})
	case key == "GET /share-networks/"+networkShareID:
		if f.subnets == 0 || (f.share == nil && f.subnets == 1) {
			f.notFound(w)
			return
		}
		subnets := make([]map[string]any, 0, f.subnets)
		for range f.subnets {
			subnets = append(subnets, map[string]any{"neutron_net_id": networkID, "neutron_subnet_id": subnetID})
		}
		f.reply(w, http.StatusOK, map[string]any{"share_network": map[string]any{
			"id": networkShareID, "share_network_subnets": subnets,
		}})
	default:
		f.t.Errorf("unexpected request %s", key)
		w.WriteHeader(http.StatusTeapot)
	}
}

func (f *fakeAPI) reply(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	assert.NoError(f.t, json.NewEncoder(w).Encode(value))
}

func (f *fakeAPI) notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = io.WriteString(w, `{"itemNotFound": {"code": 404, "message": "Share could not be found."}}`)
}

// create stores the share with the requested metadata, unless rejectCreate is
// set, and answers as the fork does or with lostCreate.
func (f *fakeAPI) create(w http.ResponseWriter, body []byte) {
	if f.rejectCreate != nil {
		f.share = nil
		w.WriteHeader(f.rejectCreate.status)
		_, _ = io.WriteString(w, f.rejectCreate.body)

		return
	}
	var created struct {
		Metadata map[string]string `json:"metadata"`
	}
	require.NoError(f.t, json.Unmarshal(body, &created))
	for k, v := range created.Metadata {
		f.share["metadata"].(map[string]any)[k] = v
	}
	f.share["status"] = "creating"
	if f.lostCreate != nil {
		w.WriteHeader(f.lostCreate.status)
		_, _ = io.WriteString(w, f.lostCreate.body)

		return
	}
	f.reply(w, http.StatusOK, map[string]any{"share": f.share})
}

// list answers the detailed share list filtered by name, status, mount path
// and metadata pairs, as the API does.
func (f *fakeAPI) list(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	var pairs map[string]string
	if raw := query.Get("metadata"); raw != "" {
		require.NoError(f.t, json.Unmarshal([]byte(raw), &pairs))
	}
	shares := []any{}
	if f.hiddenLists > 0 {
		f.hiddenLists--
		f.reply(w, http.StatusOK, map[string]any{"shares": shares})

		return
	}
	for _, candidate := range append([]map[string]any{f.share}, f.others...) {
		if candidate == nil || !listed(candidate, query.Get("name"), query.Get("status"), pairs) {
			continue
		}
		// Only the main share has mount paths.
		if path := query.Get("export_location_path"); path != "" && (candidate["id"] != shareID || f.paths[0]["path"] != path) {
			continue
		}
		shares = append(shares, candidate)
	}
	f.reply(w, http.StatusOK, map[string]any{"shares": shares})
}

func listed(candidate map[string]any, name, status string, pairs map[string]string) bool {
	if (name != "" && candidate["name"] != name) || (status != "" && candidate["status"] != status) {
		return false
	}
	metadata := candidate["metadata"].(map[string]any)
	for key, value := range pairs {
		if metadata[key] != value {
			return false
		}
	}

	return true
}

// requestQuery is a query parameter of the last request of the given kind.
func (f *fakeAPI) requestQuery(request, name string) string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.queries[request].Get(name)
}
