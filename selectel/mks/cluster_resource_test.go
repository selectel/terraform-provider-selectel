package mks

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kubeOptionsObject builds the kubernetes_options object with the given
// overrides on top of the settled options.
func kubeOptionsObject(overrides map[string]attr.Value) types.Object {
	values := map[string]attr.Value{
		"feature_gates":         types.SetValueMust(types.StringType, []attr.Value{types.StringValue("TopologyAwareHints")}),
		"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("NodeRestriction")}),
		"audit_logs": types.ObjectValueMust(auditLogsAttrTypes, map[string]attr.Value{
			"enabled":     types.BoolValue(false),
			"secret_name": types.StringNull(),
		}),
		"oidc":                 oidcObject(),
		"x509_ca_certificates": types.StringNull(),
	}
	maps.Copy(values, overrides)

	return types.ObjectValueMust(kubeOptionsAttrTypes, values)
}

// bodyFields checks fields of a request body, or of its "cluster" object when
// it has one, and that the absent ones were not sent.
func bodyFields(t *testing.T, body map[string]any, want map[string]any, absent ...string) {
	t.Helper()
	if cluster, ok := body["cluster"].(map[string]any); ok {
		body = cluster
	}
	for key, value := range want {
		got, ok := body[key]
		require.True(t, ok, "field %s missing from %v", key, body)
		assert.Equal(t, value, got, "field %s", key)
	}
	for _, key := range absent {
		_, ok := body[key]
		assert.False(t, ok, "field %s must be omitted", key)
	}
}

// oidcOf returns the oidc object of the kubernetes_options a request body
// carries.
func oidcOf(body map[string]any) map[string]any {
	cluster, _ := body["cluster"].(map[string]any)
	options, _ := cluster["kubernetes_options"].(map[string]any)
	oidc, _ := options["oidc"].(map[string]any)

	return oidc
}

// enabledOIDCObject is the OIDC object of an enabled configuration with
// only its required parameters set.
func enabledOIDCObject() types.Object {
	return types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
		"enabled":        types.BoolValue(true),
		"provider_name":  types.StringValue("keycloak"),
		"issuer_url":     types.StringValue("https://issuer.example.com"),
		"client_id":      types.StringValue("kubernetes"),
		"username_claim": types.StringValue("sub"),
		"groups_claim":   types.StringValue("groups"),
		"ca_certs":       trimmedStringValue{StringValue: types.StringNull()},
	})
}

func TestClusterCreateSendsFields(t *testing.T) {
	fake := newFakeAPI(t)
	fake.seedTask(testClusterID, "ng-1", "RESIZE_NODEGROUP", true)
	r := fake.resource()

	plan := clusterPlan()
	plan.Name = types.StringValue("Test-Cluster")
	// An unconfigured attribute carries null, not unknown, in a create plan.
	plan.CloudSubnetCIDR = types.StringNull()
	plan.ClusterType = types.StringValue("HIGH_AVAILABILITY")
	plan.WorkersType = types.StringValue("CLOUD")
	plan.CNIType = types.StringValue("CILIUM")
	plan.CNICiliumSettings = types.ObjectValueMust(ciliumAttrTypes, map[string]attr.Value{
		"envoy_daemonset": types.BoolValue(true),
		"hubble_relay":    types.BoolValue(false),
	})
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"audit_logs": types.ObjectValueMust(auditLogsAttrTypes, map[string]attr.Value{
			"enabled":     types.BoolNull(),
			"secret_name": types.StringNull(),
		}),
		"oidc": enabledOIDCObject(),
	})

	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	got := readClusterModel(t, state)

	// The create body carries what the API defaults and what expandClusterPatch
	// sends, and the nodegroup task never stalls the cluster create.
	body := fake.lastBody(routeCreateCluster)
	bodyFields(t, body, map[string]any{
		"network_type": "STANDARD", "basic": false, "cni_type": "CILIUM",
	})
	oidc := oidcOf(body)
	assert.Equal(t, map[string]any{
		"enabled": true, "provider_name": "keycloak", "issuer_url": "https://issuer.example.com",
		"client_id": "kubernetes", "username_claim": "sub", "groups_claim": "groups",
	}, oidc)

	assert.Equal(t, types.StringValue("Test-Cluster"), got.Name)
	assert.Equal(t, types.StringValue(testClusterID), got.ID)
	assert.Equal(t, types.StringValue("ACTIVE"), got.Status)
	assert.Equal(t, types.StringValue("192.0.2.10"), got.KubeAPIIP)
	assert.Equal(t, types.StringValue("fake-network"), got.NetworkID)
	assert.Equal(t, types.StringValue("fake-subnet"), got.SubnetID)
	// cloud_subnet_cidr is never returned and was planned unknown: the
	// state takes it null until the resource writes it.
	assert.True(t, got.CloudSubnetCIDR.IsNull())

	// The fake stored the project dashed, the state keeps the configured one.
	assert.Equal(t, types.StringValue(testProject), got.ProjectID)
	assert.Equal(t, 1, fake.callCount(routeCreateCluster))
}

func TestClusterProjectIDForms(t *testing.T) {
	keystone := keystoneProject
	for _, tt := range []struct {
		name       string
		planProj   string
		stateProj  string
		wantConfig string
	}{
		{name: "un-dashed configured", planProj: keystone, wantConfig: keystone},
		{name: "dashed configured", planProj: apiProject, wantConfig: apiProject},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t)
			r := fake.resource()

			plan := clusterPlan()
			plan.ProjectID = types.StringValue(tt.planProj)
			state, diags := createCluster(t, r, plan)
			require.False(t, diags.HasError(), "%v", diags)
			got := readClusterModel(t, state)
			assert.Equal(t, types.StringValue(tt.wantConfig), got.ProjectID)
			// The fake stored it dashed.
			assert.Equal(t, apiProject, fake.clusterProject(testClusterID))
			// The client got the un-dashed form.
			cfg := r.config.(*testProviderConfig)
			require.NotEmpty(t, cfg.gotProject)
			assert.Equal(t, keystone, cfg.gotProject[len(cfg.gotProject)-1])
		})
	}
}

func TestClusterWorkersType(t *testing.T) {
	for _, tt := range []struct {
		workersType string
		networkType string
	}{
		{workersType: "CLOUD", networkType: "STANDARD"},
		{workersType: "DEDICATED", networkType: "L3VPN"},
	} {
		t.Run(tt.workersType, func(t *testing.T) {
			fake := newFakeAPI(t)
			r := fake.resource()

			plan := clusterPlan()
			plan.WorkersType = types.StringValue(tt.workersType)
			_, diags := createCluster(t, r, plan)
			require.False(t, diags.HasError(), "%v", diags)

			bodyFields(t, fake.lastBody(routeCreateCluster), map[string]any{"network_type": tt.networkType})
		})
	}
}

func TestClusterCreateTaskError(t *testing.T) {
	fake := newFakeAPI(t)
	fake.failTasks("CREATE_CLUSTER", true)
	r := fake.resource()

	_, diags := createCluster(t, r, clusterPlan())
	assert.True(t, diags.HasError())
	assert.Contains(t, firstError(t, diags),
		"task CREATE_CLUSTER task-1 of cluster "+testClusterID+" ended in ERROR: fake failure (code 42): fake details of CREATE_CLUSTER")
}

// The v2.0.0 SDK keeps only the JSON GenericError message: a plain-text body
// no longer reaches the detail, unlike the fork C tested.
func TestClusterCreateUndeclaredStatus(t *testing.T) {
	for _, tt := range []struct {
		name        string
		status      int
		contentType string
		body        string
		wantErr     string
	}{
		{
			name:        "400 with the API error",
			status:      http.StatusBadRequest,
			contentType: "application/json",
			body:        `{"error":{"message":"invalid kube_version: 1.31.8"}}`,
			wantErr:     "error creating cluster: 400 Bad Request: invalid kube_version: 1.31.8",
		},
		{
			name:        "403 with a JSON error",
			status:      http.StatusForbidden,
			contentType: "application/json",
			body:        `{"error":{"message":"access denied for the project"}}`,
			wantErr:     "error creating cluster: 403 Forbidden: access denied for the project",
		},
		{
			name:        "422 with plain text",
			status:      http.StatusUnprocessableEntity,
			contentType: "text/plain",
			body:        "kube_version 1.31.8 is not supported\n",
			wantErr:     "error creating cluster: 422 Unprocessable Entity",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t)
			fake.failWithBody(routeCreateCluster, tt.status, tt.contentType, tt.body)
			r := fake.resource()

			_, diags := createCluster(t, r, clusterPlan())
			assert.True(t, diags.HasError())
			assert.Contains(t, firstError(t, diags), tt.wantErr)
		})
	}
}

// The API replaces kubernetes_options on PATCH, so the whole object goes.
func TestClusterKubernetesOptionsSentWhole(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	state, createStateDiags := createCluster(t, r, clusterPlan())
	require.False(t, createStateDiags.HasError(), "%v", createStateDiags)
	prior := readClusterModel(t, state)
	plan := prior
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("NodeRestriction"), types.StringValue("PodNodeSelector"),
		}),
		"audit_logs": types.ObjectValueMust(auditLogsAttrTypes, map[string]attr.Value{
			"enabled": types.BoolValue(true), "secret_name": types.StringValue("siem"),
		}),
		"oidc": enabledOIDCObject(),
	})

	_, diags := updateCluster(t, r, prior, plan)
	require.False(t, diags.HasError(), "%v", diags)
	body := fake.lastBody(routePatchCluster)
	cluster, _ := body["cluster"].(map[string]any)
	require.Len(t, cluster, 1)
	options, _ := cluster["kubernetes_options"].(map[string]any)
	assert.Equal(t, "[TopologyAwareHints]", fmt.Sprint(options["feature_gates"]))
	audit, _ := options["audit_logs"].(map[string]any)
	assert.Equal(t, true, audit["enabled"])
	assert.Equal(t, "siem", audit["secret_name"])
	oidc, _ := options["oidc"].(map[string]any)
	assert.Equal(t, true, oidc["enabled"])
	assert.Equal(t, "keycloak", oidc["provider_name"])
	assert.Greater(t, fake.callCount(routeTask), 0, "the provider must poll the UPGRADE_MASTERS_CONFIG task")
}

// TestClusterAutoUpgradeRecorded covers enable_patch_version_auto_upgrade:
// explicit false survives on create and changes as sent on PATCH.
func TestClusterAutoUpgradeFalse(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.EnablePatchVersionAutoUpgrade = types.BoolValue(false)
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	got := readClusterModel(t, state)
	assert.Equal(t, types.BoolValue(false), got.EnablePatchVersionAutoUpgrade)
	bodyFields(t, fake.lastBody(routeCreateCluster), map[string]any{"enable_patch_version_auto_upgrade": false})

	// Enabling it PATCHes true, and disabling it back PATCHes false.
	onPlan := got
	onPlan.EnablePatchVersionAutoUpgrade = types.BoolValue(true)
	got2, diags := updateCluster(t, r, got, onPlan)
	require.False(t, diags.HasError(), "%v", diags)
	bodyFields(t, fake.lastBody(routePatchCluster), map[string]any{"enable_patch_version_auto_upgrade": true})

	offPlan := got2
	offPlan.EnablePatchVersionAutoUpgrade = types.BoolValue(false)
	_, diags = updateCluster(t, r, got2, offPlan)
	require.False(t, diags.HasError(), "%v", diags)
	bodyFields(t, fake.lastBody(routePatchCluster), map[string]any{"enable_patch_version_auto_upgrade": false})
}

// An explicit enable_patch_version_auto_upgrade is sent on create: a BASIC
// cluster does not default it on, so false goes in the body.
func TestClusterBasicClusterAutoUpgrade(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.ClusterType = types.StringValue("BASIC")
	plan.EnablePatchVersionAutoUpgrade = types.BoolValue(false)
	_, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	bodyFields(t, fake.lastBody(routeCreateCluster), map[string]any{
		"cluster_type": "BASIC", "basic": false, "enable_patch_version_auto_upgrade": false,
	})
}

func TestClusterKubeVersionUpgrade(t *testing.T) {
	fake := newFakeAPI(t)
	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new(testKubeVersionOld)},
		mksclient.KubeVersionInfo{Version: new(testKubeVersion)},
		mksclient.KubeVersionInfo{Version: new(testKubeVersionUp)},
	)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubeVersion = types.StringValue(testKubeVersionOld)

	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	prior := readClusterModel(t, state)
	assert.Equal(t, 0, fake.callCount(routeUpgradePatch))
	assert.Equal(t, 0, fake.callCount(routeUpgradeMinor))

	// A non-latest patch within the same minor is refused.
	badPlan := prior
	badPlan.KubeVersion = types.StringValue("1.31.7")
	_, diags = updateCluster(t, r, prior, badPlan)
	assert.True(t, diags.HasError())
	assert.Contains(t, firstError(t, diags), fmt.Sprintf(
		"current version %s can't be upgraded to version 1.31.7, the latest available patch version is: %s",
		testKubeVersionOld, testKubeVersion))

	// The latest patch of the same minor goes through UPGRADE_PATCH_VERSION.
	okPlan := prior
	okPlan.KubeVersion = types.StringValue(testKubeVersion)
	got, diags := updateCluster(t, r, prior, okPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersion), got.KubeVersion)
	assert.Equal(t, 1, fake.callCount(routeUpgradePatch))
	assert.Equal(t, 0, fake.callCount(routeUpgradeMinor))

	// The next minor goes through UPGRADE_MINOR_VERSION.
	minorPlan := got
	minorPlan.KubeVersion = types.StringValue(testKubeVersionUp)
	got2, diags := updateCluster(t, r, got, minorPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersionUp), got2.KubeVersion)
	assert.Equal(t, 1, fake.callCount(routeUpgradeMinor))

	// Patch auto-upgrade moved the cluster past the configuration: the
	// refresh keeps the configured version when it is a prior patch of the
	// same minor.
	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) {
		c.KubeVersion = "1.32.5"
	})
	state2, diags := readCluster(t, r, got2)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersionUp), readClusterModel(t, state2).KubeVersion)
}

// A minor upgrade outside Terraform is read as is, so the next one-minor
// upgrade is checked against the version the cluster runs.
func TestClusterKubeVersionUpgradedOutside(t *testing.T) {
	fake := newFakeAPI(t)
	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new(testKubeVersion)},
		mksclient.KubeVersionInfo{Version: new("1.32.0")},
		mksclient.KubeVersionInfo{Version: new("1.33.1")},
	)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubeVersion = types.StringValue(testKubeVersion)
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	prior := readClusterModel(t, state)

	// Upgraded in the panel: the refresh takes the new version as is.
	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) {
		c.KubeVersion = "1.32.0"
	})
	refreshed, diags := readCluster(t, r, prior)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("1.32.0"), readClusterModel(t, refreshed).KubeVersion)

	// The upgrade starts from the version the cluster runs: the minor jump
	// lands on the latest patch of 1.33.
	afterRefresh := readClusterModel(t, refreshed)
	upPlan := afterRefresh
	upPlan.KubeVersion = types.StringValue("1.33.1")
	upState, diags := updateCluster(t, r, afterRefresh, upPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("1.33.1"), upState.KubeVersion)
	assert.Equal(t, 1, fake.callCount(routeUpgradeMinor))
}

func TestClusterImport(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()
	config := r.config.(*testProviderConfig)
	config.pool = "ru-2"

	resp := importCluster(t, r)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	got := readClusterModel(t, resp.State)
	assert.Equal(t, types.StringValue(testClusterID), got.ID)
	// The state takes the project the way the provider stores it: no dashes.
	assert.Equal(t, types.StringValue(keystoneProject), got.ProjectID)
	assert.Equal(t, types.StringValue("ru-2"), got.Pool)
}

func TestClusterImportNeedsProviderConfig(t *testing.T) {
	for _, tt := range []struct {
		name   string
		config testProviderConfig
		want   string
	}{
		{name: "no project", config: testProviderConfig{pool: "ru-1"}, want: "INFRA_PROJECT_ID must be set for the resource import"},
		{name: "no pool", config: testProviderConfig{project: testProject}, want: "INFRA_REGION must be set for the resource import"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := &clusterResource{connection: connection{config: &tt.config}}
			empty := clusterStateOf(t, r, clusterState())
			empty.Raw = empty.Raw.Copy()
			resp := resource.ImportStateResponse{State: empty}
			r.ImportState(context.Background(), resource.ImportStateRequest{ID: testClusterID}, &resp)

			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tt.want)
		})
	}
}

func TestClusterImportDashedProject(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.ProjectID = types.StringNull()
	_, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	// The provider project is dashed: the state takes it without dashes.
	importedResp := importCluster(t, r)
	require.False(t, importedResp.Diagnostics.HasError(), "%v", importedResp.Diagnostics)
	got := readClusterModel(t, importedResp.State)
	// The state takes the provider project without dashes.
	assert.Equal(t, types.StringValue(keystoneProject), got.ProjectID)

	// The config still keeps the dashed form from an import.
	state2, diags := readCluster(t, r, got)
	require.False(t, diags.HasError(), "%v", diags)
	kube := readClusterModel(t, state2)
	assert.Equal(t, types.StringValue(testClusterID), kube.ID)
}

func TestClusterReadRemovedOutside(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	fake.removeCluster(testClusterID)

	state2, diags := readCluster(t, r, readClusterModel(t, state))
	require.False(t, diags.HasError(), "%v", diags)
	assert.True(t, state2.Raw.IsNull())
}

func TestClusterDeleteTaskError(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubeVersion = types.StringValue(testKubeVersion)
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	fake.failTasks("DELETE_CLUSTER", true)
	diags = deleteCluster(t, r, readClusterModel(t, state))
	assert.True(t, diags.HasError())
	assert.Contains(t, firstError(t, diags), "task DELETE_CLUSTER task-")
	assert.Contains(t, firstError(t, diags), "ended in ERROR")

	// The failed delete left the cluster in place with a status a delete
	// takes: the retry deletes it.
	fake.failTasks("DELETE_CLUSTER", false)
	diags = deleteCluster(t, r, readClusterModel(t, state))
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, 0, fake.clusterCount())
}

func TestTaskWaiterScope(t *testing.T) {
	for _, tt := range []struct {
		name        string
		nodegroupID string
		wantTimeout bool
	}{
		{name: "cluster scope waits for cluster tasks only"},
		{name: "nodegroup scope waits for its nodegroup tasks only", nodegroupID: "ng-a"},
		{name: "a stuck task in scope times out", nodegroupID: "ng-stuck", wantTimeout: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t)
			client := fake.client()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			fake.seedCluster(testClusterID)
			fake.seedTask(testClusterID, "", "OLD", true)
			fake.seedTask(testClusterID, "ng-a", "OLD", true)
			waiter, err := newTaskWaiter(ctx, client, testClusterID, tt.nodegroupID)
			require.NoError(t, err)

			for _, nodegroupID := range []string{"", "ng-a", "ng-b", "ng-stuck"} {
				stuck := nodegroupID != tt.nodegroupID || nodegroupID == "ng-stuck"
				fake.seedTask(testClusterID, nodegroupID, "NEW", stuck)
			}

			err = waiter.Wait(ctx)
			if tt.wantTimeout {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClusterKubeVersionKeepsState(t *testing.T) {
	for _, tt := range []struct {
		name   string
		prior  string
		actual string
		want   string
	}{
		{name: "no prior, after import", prior: "", actual: "1.30.5", want: "1.30.5"},
		{name: "same version", prior: "1.30.3", actual: "1.30.3", want: "1.30.3"},
		{name: "newer patch keeps the configured one", prior: "1.30.3", actual: "1.30.5", want: "1.30.3"},
		{name: "older patch is drift", prior: "1.30.5", actual: "1.30.3", want: "1.30.3"},
		{name: "newer minor is drift", prior: "1.30.5", actual: "1.31.1", want: "1.31.1"},
		{name: "newer major is drift", prior: "1.30.5", actual: "2.0.1", want: "2.0.1"},
		{name: "older minor is drift", prior: "1.31.1", actual: "1.30.5", want: "1.30.5"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, kubeVersionKeepsState(tt.prior, tt.actual))
		})
	}
}

func TestClusterCreateTimeout(t *testing.T) {
	for _, tt := range []struct {
		name      string
		readFails bool
		wantError string
	}{
		{name: "read after the wait succeeds", wantError: "Error waiting for the cluster to become ready"},
		{name: "read after the wait fails", readFails: true, wantError: "Error reading cluster"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t)
			fake.stickTasks("CREATE_CLUSTER", true)
			if tt.readFails {
				fake.fail(routeCluster, http.StatusInternalServerError)
			}
			r := fake.resource()

			plan := clusterPlan()
			plan.Timeouts = shortTimeouts()
			plan.Timeouts.Object = types.ObjectValueMust(
				map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType},
				map[string]attr.Value{
					"create": types.StringValue("1s"), "update": types.StringValue("2s"), "delete": types.StringValue("2s"),
				},
			)
			state, diags := createCluster(t, r, plan)
			assert.True(t, diags.HasError())
			assert.Contains(t, errorText(t, diags), tt.wantError)
			assert.Contains(t, errorText(t, diags), "context deadline exceeded")

			// The timed-out cluster was saved in state: a retry-able delete
			// later replaces it once the task unsticks.
			assert.Equal(t, 1, fake.callCount(routeCreateCluster))
			fake.stickTasks("CREATE_CLUSTER", false)
			fake.settleAfter(testClusterID, 3, "ACTIVE")
			fake.fail(routeCluster, 0)
			diags = deleteCluster(t, r, readClusterModel(t, state))
			assert.False(t, diags.HasError(), "%v", diags)
		})
	}
}

func TestClusterDeleteWaitsForStatus(t *testing.T) {
	for _, tt := range []struct {
		name    string
		status  string
		settle  *string
		wantErr bool
	}{
		{name: "pending create ends", status: "PENDING_CREATE", settle: new("ACTIVE")},
		{name: "error deletes at once", status: "ERROR"},
		{name: "maintenance deletes at once", status: "MAINTENANCE"},
		{name: "gone while waiting", status: "PENDING_CREATE", settle: new("")},
		{name: "still pending at the timeout", status: "PENDING_CREATE", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t)
			r := fake.resource()

			plan := clusterPlan()
			plan.Timeouts = shortTimeouts()
			plan.Timeouts.Object = types.ObjectValueMust(
				map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType},
				map[string]attr.Value{
					"create": types.StringValue("2s"), "update": types.StringValue("2s"), "delete": types.StringValue("1s"),
				},
			)
			state, diags := createCluster(t, r, plan)
			require.False(t, diags.HasError(), "%v", diags)

			fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) {
				c.Status = mksclient.ClusterDetailedStatus(tt.status)
			})
			if tt.settle != nil {
				fake.settleAfter(testClusterID, 3, *tt.settle)
			}

			diags = deleteCluster(t, r, readClusterModel(t, state))
			if tt.wantErr {
				assert.True(t, diags.HasError())
				assert.Contains(t, errorText(t, diags), "cluster status PENDING_CREATE does not allow the delete yet")
				assert.Contains(t, errorText(t, diags), "context deadline exceeded")
				assert.Equal(t, 0, fake.callCount(routeDeleteCluster))

				// The cluster stays in state; a retry deletes it once ACTIVE.
				fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "ACTIVE" })
				diags = deleteCluster(t, r, readClusterModel(t, state))
				assert.False(t, diags.HasError(), "%v", diags)
			}
			assert.Equal(t, 0, fake.clusterCount())
			if !tt.wantErr {
				assert.Equal(t, 0, fake.clusterCount())
			}
		})
	}
}

func TestClusterUpdateTimeout(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.Timeouts = shortTimeouts()
	plan.Timeouts.Object = types.ObjectValueMust(
		map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType},
		map[string]attr.Value{
			"create": types.StringValue("2s"), "update": types.StringValue("1s"), "delete": types.StringValue("2s"),
		},
	)
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	prior := readClusterModel(t, state)

	fake.stickTasks("UPGRADE_MASTERS_CONFIG", true)
	newPlan := prior
	newPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"feature_gates": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("NewFeatureGate")}),
	})
	_, diags = updateCluster(t, r, prior, newPlan)
	assert.True(t, diags.HasError())
	assert.Contains(t, errorText(t, diags), "Error updating cluster")
	assert.Contains(t, firstError(t, diags), "context deadline exceeded")

	// The API stored the options before its task: the refresh reads them
	// and the task ends on its own.
	fake.stickTasks("UPGRADE_MASTERS_CONFIG", false)
	fake.settleAfter(testClusterID, 1, "ACTIVE")
	state2, diags := readCluster(t, r, prior)
	require.False(t, diags.HasError(), "%v", diags)
	got := readClusterModel(t, state2)
	assert.Contains(t, got.KubernetesOptions.String(), "NewFeatureGate")
}

func TestClusterDeleteTimeout(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.Timeouts = shortTimeouts()
	plan.Timeouts.Object = types.ObjectValueMust(
		map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType},
		map[string]attr.Value{
			"create": types.StringValue("2s"), "update": types.StringValue("2s"), "delete": types.StringValue("1s"),
		},
	)
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	fake.stickTasks("DELETE_CLUSTER", true)
	diags = deleteCluster(t, r, readClusterModel(t, state))
	assert.True(t, diags.HasError())
	assert.Contains(t, errorText(t, diags), "Error deleting cluster")
	assert.Contains(t, firstError(t, diags), "context deadline exceeded")
	// Only one DELETE went out: the wait, not another request.
	assert.Equal(t, 1, fake.callCount(routeDeleteCluster))

	// The cluster stays PENDING_DELETE until the delete ends in the API.
	fake.settleAfter(testClusterID, 3, "")
	diags = deleteCluster(t, r, readClusterModel(t, state))
	assert.False(t, diags.HasError(), "%v", diags)
	// The second destroy waited for the first delete instead of sending
	// another one, which the API would refuse with 409.
	assert.Equal(t, 1, fake.callCount(routeDeleteCluster))
	assert.Equal(t, 0, fake.clusterCount())
}

func TestClusterX509OnCreate(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringValue("Y2VydA=="),
	})
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	got := readClusterModel(t, state)
	assert.Equal(t, 1, fake.callCount(routeCreateCluster))
	assert.Equal(t, 1, fake.callCount(routePatchCluster))

	// The create body carries no x509: the PATCH carries the whole object.
	body := fake.lastBody(routePatchCluster)
	cluster, _ := body["cluster"].(map[string]any)
	require.Len(t, cluster, 1)
	options, _ := cluster["kubernetes_options"].(map[string]any)
	assert.Equal(t, "Y2VydA==", options["x509_ca_certificates"])
	assert.Equal(t, "[TopologyAwareHints]", fmt.Sprint(options["feature_gates"]))
	assert.Equal(t, "Y2VydA==", fake.storedX509(testClusterID))

	var kubeOptions clusterKubeOptionsModel
	diags2 := got.KubernetesOptions.As(context.Background(), &kubeOptions, basetypes.ObjectAsOptions{})
	require.False(t, diags2.HasError())
	assert.Equal(t, "Y2VydA==", kubeOptions.X509CACertificates.ValueString())
	assert.Equal(t, types.StringValue("ACTIVE"), got.Status)
}

func TestClusterUpperCaseName(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.Name = types.StringValue("TF-v2")
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	got := readClusterModel(t, state)

	// The API stores tf-v2; the state keeps the configured case.
	bodyFields(t, fake.lastBody(routeCreateCluster), map[string]any{"name": "TF-v2"})
	assert.Equal(t, types.StringValue("TF-v2"), got.Name)
}

func TestClusterNameCaseOnlyChange(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.Name = types.StringValue("TF-v2")
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	prior := readClusterModel(t, state)

	// Only the case changes: fromAPI keeps the prior value, so the state is
	// consistent and the plan empty.
	changed := prior
	changed.Name = types.StringValue("tf-V2")
	got, diags := updateCluster(t, r, prior, changed)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("tf-V2"), got.Name)
}

func TestClusterX509OnCreateFails(t *testing.T) {
	fake := newFakeAPI(t)
	fake.fail(routePatchCluster, http.StatusConflict)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringValue("Y2VydA=="),
	})
	state, diags := createCluster(t, r, plan)
	// The create succeeds with a warning, the resource is not tainted.
	require.False(t, diags.HasError(), "%v", diags)
	assert.True(t, diags.WarningsCount() > 0)
	got := readClusterModel(t, state)
	assert.Equal(t, types.StringValue(testClusterID), got.ID)
	assert.Equal(t, 1, fake.callCount(routeCreateCluster))
	assert.Equal(t, 1, fake.callCount(routePatchCluster))

	fake.fail(routePatchCluster, 0)
	// The next apply only PATCHes the x509.
	prior := got
	newPlan := prior
	newPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringValue("Y2VydA=="),
	})
	got2, diags := updateCluster(t, r, prior, newPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, 0, fake.callCount(routeDeleteCluster))
	assert.Equal(t, 2, fake.callCount(routePatchCluster))
	body := fake.lastBody(routePatchCluster)
	cluster, _ := body["cluster"].(map[string]any)
	require.Len(t, cluster, 1)
	options, _ := cluster["kubernetes_options"].(map[string]any)
	assert.Equal(t, "Y2VydA==", options["x509_ca_certificates"])
	assert.Equal(t, "Y2VydA==", fake.storedX509(testClusterID))
	var kubeOptions clusterKubeOptionsModel
	diags2 := got2.KubernetesOptions.As(context.Background(), &kubeOptions, basetypes.ObjectAsOptions{})
	require.False(t, diags2.HasError())
	assert.Equal(t, "Y2VydA==", kubeOptions.X509CACertificates.ValueString())
}

func TestClusterOIDCCACertsWhitespace(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
			"enabled":        types.BoolValue(true),
			"provider_name":  types.StringValue("keycloak"),
			"issuer_url":     types.StringValue("https://issuer.example.com"),
			"client_id":      types.StringValue("kubernetes"),
			"username_claim": types.StringValue("sub"),
			"groups_claim":   types.StringValue("groups"),
			"ca_certs":       trimmedString("cert-1\n"),
		}),
	})
	_, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	// The plan carried the trailing newline, the API stores what it trimmed:
	// like file("ca.pem"), the state keeps the configured form until the
	// next apply converges it.
	body := oidcOf(fake.lastBody(routeCreateCluster))
	assert.Equal(t, "cert-1\n", body["ca_certs"])
	fake.withinCluster(testClusterID, func(c mksclient.ClusterDetailed) {
		assert.Equal(t, "cert-1", c.KubernetesOptions.Oidc.CaCerts)
	})
}

func TestClusterOIDCClaimDefaults(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	// The claims omitted plan the API defaults, so the apply is consistent
	// with what the API stores.
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": enabledOIDCObject(),
	})

	// The plan-time modifier plans the API defaults.
	statePlan := tfsdk.Plan(clusterStateOf(t, r, plan))
	modifyResp := resource.ModifyPlanResponse{Plan: statePlan}
	modifyReq := resource.ModifyPlanRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, plan)),
		Plan:   statePlan, State: clusterStateOf(t, r, clusterState()),
	}
	r.ModifyPlan(context.Background(), modifyReq, &modifyResp)
	require.False(t, modifyResp.Diagnostics.HasError(), "%v", modifyResp.Diagnostics)

	// The plan modifiers themselves run only inside the framework planner, so
	// the claim defaults are checked on the modifier directly.
	modifierReq := func(name string) planmodifier.StringRequest {
		return planmodifier.StringRequest{
			Config:      tfsdk.Config(clusterStateOf(t, r, plan)),
			ConfigValue: types.StringNull(),
			StateValue:  types.StringNull(),
			Path:        path.Root("kubernetes_options").AtName("oidc").AtName(name),
		}
	}
	for name, want := range map[string]string{"username_claim": "sub", "groups_claim": "groups"} {
		modResp := planmodifier.StringResponse{PlanValue: types.StringNull()}
		oidcPlanModifier{enabledDefault: want}.PlanModifyString(context.Background(), modifierReq(name), &modResp)
		require.False(t, modResp.Diagnostics.HasError(), "%v", modResp.Diagnostics)
		assert.Equal(t, types.StringValue(want), modResp.PlanValue, name)
	}

	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	got := readClusterModel(t, state)
	oidc := oidcOf(fake.lastBody(routeCreateCluster))
	assert.Equal(t, map[string]any{
		"enabled": true, "provider_name": "keycloak", "issuer_url": "https://issuer.example.com",
		"client_id": "kubernetes", "username_claim": "sub", "groups_claim": "groups",
	}, oidc)
	var kubeOptions clusterKubeOptionsModel
	diags2 := got.KubernetesOptions.As(context.Background(), &kubeOptions, basetypes.ObjectAsOptions{})
	require.False(t, diags2.HasError())
	assert.Contains(t, kubeOptions.OIDC.String(), `"username_claim":"sub"`)
	assert.Contains(t, kubeOptions.OIDC.String(), `"groups_claim":"groups"`)
}

func TestClusterOIDCDisable(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
			"enabled":        types.BoolValue(true),
			"provider_name":  types.StringValue("keycloak"),
			"issuer_url":     types.StringValue("https://issuer.example.com"),
			"client_id":      types.StringValue("kubernetes"),
			"username_claim": types.StringValue("sub"),
			"groups_claim":   types.StringValue("groups"),
			"ca_certs":       trimmedString("cert"),
		}),
	})
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	prior := readClusterModel(t, state)

	// Disabling it sends the parameters empty, which is the only form the
	// API accepts, and the state clears them.
	newPlan := prior
	newPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
			"enabled":        types.BoolValue(false),
			"provider_name":  types.StringNull(),
			"issuer_url":     types.StringNull(),
			"client_id":      types.StringNull(),
			"username_claim": types.StringNull(),
			"groups_claim":   types.StringNull(),
			"ca_certs":       trimmedString(""),
		}),
	})
	got, diags := updateCluster(t, r, prior, newPlan)
	require.False(t, diags.HasError(), "%v", diags)
	oidc := oidcOf(fake.lastBody(routePatchCluster))
	assert.Equal(t, map[string]any{
		"enabled": false, "provider_name": "", "issuer_url": "", "client_id": "", "username_claim": "", "groups_claim": "",
	}, oidc)
	assert.Contains(t, got.KubernetesOptions.String(), `"enabled":false`)

	// Enabling it back sends the configured parameters, and the API rejects
	// a disabled oidc with parameters.
	reEnable := prior
	_, diags = updateCluster(t, r, readClusterModel(t, state), reEnable)
	require.False(t, diags.HasError(), "%v", diags)
}

func TestClusterOIDCUnknownWhileDisabled(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": enabledOIDCObject(),
	})
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	prior := readClusterModel(t, state)

	// issuer_url is unknown until terraform_data is created: the apply-time
	// plan carries it as null in a disabled oidc.
	newPlan := prior
	newPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
			"enabled":        types.BoolValue(false),
			"provider_name":  types.StringNull(),
			"issuer_url":     types.StringUnknown(),
			"client_id":      types.StringNull(),
			"username_claim": types.StringNull(),
			"groups_claim":   types.StringNull(),
			"ca_certs":       trimmedString(""),
		}),
	})
	_, diags = updateCluster(t, r, prior, newPlan)
	require.False(t, diags.HasError(), "%v", diags)
	oidc := oidcOf(fake.lastBody(routePatchCluster))
	assert.Equal(t, map[string]any{
		"enabled": false, "provider_name": "", "issuer_url": "", "client_id": "", "username_claim": "", "groups_claim": "",
	}, oidc)
}

func TestClusterStatusOnUpdate(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.CNIType = types.StringValue("CILIUM")
	plan.CNICiliumSettings = types.ObjectValueMust(ciliumAttrTypes, map[string]attr.Value{
		"envoy_daemonset": types.BoolValue(true),
		"hubble_relay":    types.BoolValue(false),
	})
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("NodeRestriction")}),
	})
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	prior := readClusterModel(t, state)

	// The refresh sees the cluster in its maintenance window; the finished
	// task leaves it ACTIVE.
	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "MAINTENANCE" })
	newPlan := prior
	newPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("NodeRestriction"), types.StringValue("PodNodeSelector"),
		}),
	})
	got, diags := updateCluster(t, r, prior, newPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("ACTIVE"), got.Status)
	assert.Equal(t, 1, fake.callCount(routePatchCluster))

	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "MAINTENANCE" })
	ciliumPlan := got
	ciliumPlan.CNICiliumSettings = types.ObjectValueMust(ciliumAttrTypes, map[string]attr.Value{
		"envoy_daemonset": types.BoolValue(true),
		"hubble_relay":    types.BoolValue(true),
	})
	got2, diags := updateCluster(t, r, got, ciliumPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("ACTIVE"), got2.Status)
	assert.Greater(t, fake.callCount(routePatchCluster), 1)
}

func TestClusterCiliumSettingsNeedCilium(t *testing.T) {
	for _, tt := range []struct {
		name    string
		cniType types.String
	}{
		{name: "cni_type null"},
		{name: "cni_type CALICO", cniType: types.StringValue("CALICO")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t)
			r := fake.resource()
			plan := clusterPlan()
			plan.CNIType = tt.cniType
			plan.CNICiliumSettings = types.ObjectValueMust(ciliumAttrTypes, map[string]attr.Value{
				"envoy_daemonset": types.BoolValue(true),
				"hubble_relay":    types.BoolValue(false),
			})
			config := tfsdk.Config(clusterStateOf(t, r, plan))
			resp := resource.ValidateConfigResponse{}
			r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: config}, &resp)

			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, "Cilium settings need the Cilium CNI", resp.Diagnostics.Errors()[0].Summary())
		})
	}
}

func TestClusterUpdateTaskError(t *testing.T) {
	fake := newFakeAPI(t)
	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new(testKubeVersionOld)},
		mksclient.KubeVersionInfo{Version: new(testKubeVersion)},
	)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubeVersion = types.StringValue(testKubeVersionOld)
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("NodeRestriction")}),
	})
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	prior := readClusterModel(t, state)

	fake.failTasks("UPGRADE_PATCH_VERSION", true)
	upPlan := prior
	upPlan.KubeVersion = types.StringValue(testKubeVersion)
	_, diags = updateCluster(t, r, prior, upPlan)
	assert.True(t, diags.HasError())
	assert.Contains(t, firstError(t, diags), "task UPGRADE_PATCH_VERSION task-")
	assert.Contains(t, firstError(t, diags), "ended in ERROR")

	// The cluster stays in state with the prior version, so the upgrade is
	// planned again.
	fake.failTasks("UPGRADE_PATCH_VERSION", false)
	got, diags := updateCluster(t, r, readClusterModel(t, state), upPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersion), got.KubeVersion)

	// Upgraded in place, not recreated.
	assert.Equal(t, 1, fake.callCount(routeCreateCluster))
	assert.Equal(t, 2, fake.callCount(routeUpgradePatch))

	fake.failTasks("UPGRADE_MASTERS_CONFIG", true)
	mastersPlan := got
	mastersPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("NodeRestriction"), types.StringValue("PodNodeSelector"),
		}),
	})
	_, diags = updateCluster(t, r, got, mastersPlan)
	assert.True(t, diags.HasError())
	assert.Contains(t, firstError(t, diags), "task UPGRADE_MASTERS_CONFIG task-")
	assert.Contains(t, firstError(t, diags), "ended in ERROR")

	// The API stored the requested options before the task ran: the refresh
	// reads them.
	fake.failTasks("UPGRADE_MASTERS_CONFIG", false)
	state2, diags := readCluster(t, r, got)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Contains(t, readClusterModel(t, state2).KubernetesOptions.String(), "PodNodeSelector")
}

func TestTaskWaiterListsAllTasks(t *testing.T) {
	fake := newFakeAPI(t)
	fake.seedCluster(testClusterID)
	client := fake.client()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// More tasks than one page of the old paging; every old one is stuck, so
	// the wait times out if the snapshot misses any of them.
	const oldTasks = 250
	for range oldTasks {
		fake.seedTask(testClusterID, "", "OLD", true)
	}
	waiter, err := newTaskWaiter(ctx, client, testClusterID, "")
	require.NoError(t, err)
	require.Len(t, waiter.before, oldTasks)
	fake.seedTask(testClusterID, "", "NEW", false)

	err = waiter.Wait(ctx)
	require.NoError(t, err)
	// One call per listing: the snapshot and the new tasks.
	assert.Equal(t, 2, fake.callCount(routeTasks))
	assert.Contains(t, fake.lastQuery(routeTasks), "limit=0")
}
