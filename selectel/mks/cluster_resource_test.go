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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func oidcOf(body map[string]any) map[string]any {
	cluster, _ := body["cluster"].(map[string]any)
	options, _ := cluster["kubernetes_options"].(map[string]any)
	oidc, _ := options["oidc"].(map[string]any)

	return oidc
}

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
	assert.True(t, got.CloudSubnetCIDR.IsNull())

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
			assert.Equal(t, apiProject, fake.clusterProject(testClusterID))
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

	badPlan := prior
	badPlan.KubeVersion = types.StringValue("1.31.7")
	_, diags = updateCluster(t, r, prior, badPlan)
	assert.True(t, diags.HasError())
	assert.Contains(t, firstError(t, diags), fmt.Sprintf(
		"current version %s can't be upgraded to version 1.31.7, the latest available patch version is: %s",
		testKubeVersionOld, testKubeVersion))

	okPlan := prior
	okPlan.KubeVersion = types.StringValue(testKubeVersion)
	got, diags := updateCluster(t, r, prior, okPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersion), got.KubeVersion)
	assert.Equal(t, 1, fake.callCount(routeUpgradePatch))
	assert.Equal(t, 0, fake.callCount(routeUpgradeMinor))

	minorPlan := got
	minorPlan.KubeVersion = types.StringValue(testKubeVersionUp)
	got2, diags := updateCluster(t, r, got, minorPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersionUp), got2.KubeVersion)
	assert.Equal(t, 1, fake.callCount(routeUpgradeMinor))

	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) {
		c.KubeVersion = "1.32.5"
	})
	state2, diags := readCluster(t, r, got2)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersionUp), readClusterModel(t, state2).KubeVersion)
}

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

	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) {
		c.KubeVersion = "1.32.0"
	})
	refreshed, diags := readCluster(t, r, prior)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("1.32.0"), readClusterModel(t, refreshed).KubeVersion)

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

	importedResp := importCluster(t, r)
	require.False(t, importedResp.Diagnostics.HasError(), "%v", importedResp.Diagnostics)
	got := readClusterModel(t, importedResp.State)
	assert.Equal(t, types.StringValue(keystoneProject), got.ProjectID)

	state2, diags := readCluster(t, r, got)
	require.False(t, diags.HasError(), "%v", diags)
	kube := readClusterModel(t, state2)
	assert.Equal(t, types.StringValue(testClusterID), kube.ID)
	// The import resolved the project from the dashed form the API returns,
	// and the read kept it in the keystone form.
	assert.Equal(t, types.StringValue(keystoneProject), kube.ProjectID)
	// The API never returns cloud_subnet_cidr, so the imported state has none.
	assert.True(t, kube.CloudSubnetCIDR.IsNull())
	// kube_version is kept from the API even without a configured one.
	assert.Equal(t, types.StringValue(testKubeVersion), kube.KubeVersion)
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
			plan.Timeouts = clusterTimeouts(time.Second, 2*time.Second, 2*time.Second)
			state, diags := createCluster(t, r, plan)
			assert.True(t, diags.HasError())
			assert.Contains(t, errorText(t, diags), tt.wantError)
			assert.Contains(t, errorText(t, diags), "context deadline exceeded")

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
			plan.Timeouts = clusterTimeouts(2*time.Second, 2*time.Second, time.Second)
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

				fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "ACTIVE" })
				diags = deleteCluster(t, r, readClusterModel(t, state))
				assert.False(t, diags.HasError(), "%v", diags)
			}
			assert.Equal(t, 0, fake.clusterCount())
		})
	}
}

func TestClusterUpdateTimeout(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.Timeouts = clusterTimeouts(2*time.Second, time.Second, 2*time.Second)
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
	plan.Timeouts = clusterTimeouts(2*time.Second, 2*time.Second, time.Second)
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	fake.stickTasks("DELETE_CLUSTER", true)
	diags = deleteCluster(t, r, readClusterModel(t, state))
	assert.True(t, diags.HasError())
	assert.Contains(t, errorText(t, diags), "Error deleting cluster")
	assert.Contains(t, firstError(t, diags), "context deadline exceeded")
	assert.Equal(t, 1, fake.callCount(routeDeleteCluster))

	fake.settleAfter(testClusterID, 3, "")
	diags = deleteCluster(t, r, readClusterModel(t, state))
	assert.False(t, diags.HasError(), "%v", diags)
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

	changed := prior
	changed.Name = types.StringValue("tf-V2")
	got, diags := updateCluster(t, r, prior, changed)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("tf-V2"), got.Name)
}

func TestClusterResourceConfigure(t *testing.T) {
	r := &clusterResource{}
	resp := resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{}, &resp)
	assert.False(t, resp.Diagnostics.HasError())

	invalidResp := resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "not-a-config"}, &invalidResp)
	assert.True(t, invalidResp.Diagnostics.HasError())
	assert.Equal(t, "Invalid provider configuration", invalidResp.Diagnostics.Errors()[0].Summary())

	metadataResp := resource.MetadataResponse{}
	NewClusterResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "selectel"}, &metadataResp)
	assert.Equal(t, "selectel_mks_cluster_v2", metadataResp.TypeName)
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
	require.False(t, diags.HasError(), "%v", diags)
	assert.True(t, diags.WarningsCount() > 0)
	assert.Equal(t, "X509 CA certificates of the cluster are not set", diags.Warnings()[0].Summary())
	got := readClusterModel(t, state)
	assert.Equal(t, types.StringValue(testClusterID), got.ID)
	assert.Equal(t, 1, fake.callCount(routeCreateCluster))
	assert.Equal(t, 1, fake.callCount(routePatchCluster))

	fake.fail(routePatchCluster, 0)
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
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	body := oidcOf(fake.lastBody(routeCreateCluster))
	assert.Equal(t, "cert-1\n", body["ca_certs"])
	got := readClusterModel(t, state)
	var kubeOptions clusterKubeOptionsModel
	diags2 := got.KubernetesOptions.As(context.Background(), &kubeOptions, basetypes.ObjectAsOptions{})
	require.False(t, diags2.HasError())
	var oidc clusterOIDCModel
	diags3 := kubeOptions.OIDC.As(context.Background(), &oidc, basetypes.ObjectAsOptions{})
	require.False(t, diags3.HasError())
	// The API stores the trimmed cert, and semantic equality must see the
	// configured wrapped one as unchanged.
	same, diags4 := oidc.CACerts.StringSemanticEquals(context.Background(), trimmedString("cert-1"))
	require.False(t, diags4.HasError())
	assert.True(t, same, "the state value must equal a differently-wrapped identical cert")
}

func TestClusterOIDCClaimDefaults(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": enabledOIDCObject(),
	})

	statePlan := tfsdk.Plan(clusterStateOf(t, r, plan))
	modifyResp := resource.ModifyPlanResponse{Plan: statePlan}
	modifyReq := resource.ModifyPlanRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, plan)),
		Plan:   statePlan, State: clusterStateOf(t, r, clusterState()),
	}
	r.ModifyPlan(context.Background(), modifyReq, &modifyResp)
	require.False(t, modifyResp.Diagnostics.HasError(), "%v", modifyResp.Diagnostics)

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

	// The disabled-OIDC zeroing plans the empty string the API stores for an
	// omitted parameter; reverting the modifier leaves the null config.
	modReq := planmodifier.StringRequest{
		Config:      tfsdk.Config(clusterStateOf(t, r, newPlan)),
		State:       clusterStateOf(t, r, prior),
		ConfigValue: types.StringNull(),
		StateValue:  types.StringValue("https://issuer.example.com"),
		Path:        path.Root("kubernetes_options").AtName("oidc").AtName("issuer_url"),
	}
	resp := planmodifier.StringResponse{PlanValue: types.StringNull()}
	oidcPlanModifier{}.PlanModifyString(context.Background(), modReq, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.True(t, !resp.PlanValue.IsNull() && resp.PlanValue.ValueString() == "")
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

	fake.failTasks("UPGRADE_PATCH_VERSION", false)
	got, diags := updateCluster(t, r, readClusterModel(t, state), upPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersion), got.KubeVersion)

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
	assert.Equal(t, 2, fake.callCount(routeTasks))
	assert.Contains(t, fake.lastQuery(routeTasks), "limit=0")
}

func TestClusterModifyPlanKubeVersion(t *testing.T) {
	for _, tt := range []struct {
		name  string
		state string
		plan  string
	}{
		{name: "lower same minor", state: "1.31.8", plan: "1.31.5"},
		{name: "lower minor", state: "1.32.8", plan: "1.31.5"},
		{name: "lower major", state: "2.31.1", plan: "1.31.5"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := &clusterResource{}
			state := clusterState()
			state.KubeVersion = types.StringValue(tt.state)
			plan := clusterState()
			plan.KubeVersion = types.StringValue(tt.plan)
			req := resource.ModifyPlanRequest{
				Config: tfsdk.Config(clusterStateOf(t, r, plan)), Plan: tfsdk.Plan(clusterStateOf(t, r, plan)),
				State: clusterStateOf(t, r, state),
			}
			resp := resource.ModifyPlanResponse{Plan: req.Plan}
			r.ModifyPlan(context.Background(), req, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var got clusterModel
			require.False(t, resp.Plan.Get(context.Background(), &got).HasError())
			assert.Equal(t, types.StringValue(tt.state), got.KubeVersion)
		})
	}
}

func TestClusterValidateKubeVersionChange(t *testing.T) {
	for _, tt := range []struct {
		name    string
		current string
		desired string
		wantErr string
	}{
		{name: "skipped minor", current: "1.31.8", desired: "1.33.1", wantErr: "must be upgraded one minor version at a time"},
		{name: "major change", current: "1.31.8", desired: "2.31.1", wantErr: "can't be upgraded to version"},
		{name: "downgrade", current: "1.32.8", desired: "1.31.5", wantErr: "can't be downgraded to version"},
		{name: "patch downgrade", current: "1.31.8", desired: "1.31.5", wantErr: "can't be downgraded to version"},
		{name: "one minor up", current: "1.31.8", desired: "1.32.3"},
		{name: "patch up", current: "1.31.8", desired: "1.31.9"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKubeVersionChange(tt.current, tt.desired)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestClusterBasicClusterAutoUpgradeRejectedAtPlan(t *testing.T) {
	r := &clusterResource{}
	config := clusterPlan()
	config.ClusterType = types.StringValue("BASIC")
	config.EnablePatchVersionAutoUpgrade = types.BoolValue(true)
	plan := tfsdk.Config(clusterStateOf(t, r, config))
	resp := resource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: plan}, &resp)
	require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, "Patch version auto-upgrade is not available for BASIC clusters", resp.Diagnostics.Errors()[0].Summary())
}

func TestClusterOIDCModifierRules(t *testing.T) {
	r := &clusterResource{}
	oidc := func(enabled bool) types.Object {
		return types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
			"enabled":        types.BoolValue(enabled),
			"provider_name":  types.StringValue("keycloak"),
			"issuer_url":     types.StringValue("https://issuer.example.com"),
			"client_id":      types.StringValue("kubernetes"),
			"username_claim": types.StringNull(),
			"groups_claim":   types.StringNull(),
			"ca_certs":       trimmedStringValue{StringValue: types.StringNull()},
		})
	}
	t.Run("OIDC parameter set while OIDC is disabled", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidc(false)})
		modReq := planmodifier.StringRequest{
			Config:      tfsdk.Config(clusterStateOf(t, r, config)),
			ConfigValue: types.StringValue("keycloak"),
			StateValue:  types.StringNull(),
			Path:        path.Root("kubernetes_options").AtName("oidc").AtName("provider_name"),
		}
		resp := planmodifier.StringResponse{PlanValue: types.StringValue("keycloak")}
		oidcPlanModifier{}.PlanModifyString(context.Background(), modReq, &resp)
		require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, "OIDC parameter set while OIDC is disabled", resp.Diagnostics.Errors()[0].Summary())
		assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "oidc parameters cannot be configured when it is disabled")
	})
	for _, name := range []string{"username_claim", "groups_claim"} {
		t.Run(name+" cannot be empty while enabled", func(t *testing.T) {
			enabled := oidc(true)
			enabledObj := clusterPlan()
			enabledObj.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": enabled})
			modReq := planmodifier.StringRequest{
				Config:      tfsdk.Config(clusterStateOf(t, r, enabledObj)),
				ConfigValue: types.StringValue(""),
				StateValue:  types.StringNull(),
				Path:        path.Root("kubernetes_options").AtName("oidc").AtName(name),
			}
			resp := planmodifier.StringResponse{PlanValue: types.StringValue("")}
			oidcPlanModifier{enabledDefault: "sub"}.PlanModifyString(context.Background(), modReq, &resp)
			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, "Empty OIDC claim while OIDC is enabled", resp.Diagnostics.Errors()[0].Summary())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), name+" cannot be empty while OIDC is enabled: omit it to get")
		})
	}
}

func TestClusterOIDCKeepsStateWhileEnabled(t *testing.T) {
	r := &clusterResource{}
	plan := clusterPlan()
	enabledOIDC := types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
		"enabled":        types.BoolValue(true),
		"provider_name":  types.StringNull(),
		"issuer_url":     types.StringNull(),
		"client_id":      types.StringNull(),
		"username_claim": types.StringNull(),
		"groups_claim":   types.StringNull(),
		"ca_certs":       trimmedStringValue{StringValue: types.StringNull()},
	})
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": enabledOIDC})
	modReq := planmodifier.StringRequest{
		Config:      tfsdk.Config(clusterStateOf(t, r, plan)),
		State:       clusterStateOf(t, r, plan),
		StateValue:  types.StringValue("keycloak"),
		ConfigValue: types.StringNull(),
		Path:        path.Root("kubernetes_options").AtName("oidc").AtName("provider_name"),
	}
	chainReq := modReq
	chainReq.PlanValue = types.StringUnknown()
	keepResp := planmodifier.StringResponse{PlanValue: chainReq.PlanValue}
	stringplanmodifier.UseStateForUnknown().PlanModifyString(context.Background(), chainReq, &keepResp)
	resp := planmodifier.StringResponse{PlanValue: keepResp.PlanValue}
	oidcPlanModifier{}.PlanModifyString(context.Background(), modReq, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, types.StringValue("keycloak"), resp.PlanValue)
}

func TestClusterModifyPlanSetsUnknowns(t *testing.T) {
	r := &clusterResource{}
	changePlan := clusterState()
	changePlan.KubeVersion = types.StringValue(testKubeVersionUp)
	changePlan.MaintenanceWindowStart = types.StringValue("07:00:00")
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, changePlan)),
		Plan:   tfsdk.Plan(clusterStateOf(t, r, changePlan)), State: clusterStateOf(t, r, clusterState()),
	}
	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(context.Background(), req, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got clusterModel
	require.False(t, resp.Plan.Get(context.Background(), &got).HasError())
	assert.True(t, got.Status.IsUnknown())
	assert.True(t, got.MaintenanceWindowEnd.IsUnknown())
}

func TestClusterNameCaseModifier(t *testing.T) {
	for _, tt := range []struct {
		name     string
		plan     types.String
		expected types.String
	}{
		{name: "case-only change keeps the state name", plan: types.StringValue("Test-Cluster"), expected: types.StringValue("test-cluster")},
		{name: "a different name is not rewritten", plan: types.StringValue("other-cluster"), expected: types.StringValue("other-cluster")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := planmodifier.StringResponse{PlanValue: tt.plan}
			nameCase{}.PlanModifyString(context.Background(), planmodifier.StringRequest{
				StateValue: types.StringValue("test-cluster"), PlanValue: tt.plan,
			}, &resp)
			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}

func TestClusterSameProjectIDModifier(t *testing.T) {
	for _, tt := range []struct {
		name     string
		plan     types.String
		expected types.String
	}{
		{name: "dashed form keeps the state project", plan: types.StringValue("a07abc12-3105-46f1-b929-1ab301a7d75b"), expected: types.StringValue("a07abc12310546f1b9291ab301a7d75b")},
		{name: "another project is not rewritten", plan: types.StringValue("c07abc12310546f1b9291ab301a7d75b"), expected: types.StringValue("c07abc12310546f1b9291ab301a7d75b")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := planmodifier.StringResponse{PlanValue: tt.plan}
			sameProjectID{}.PlanModifyString(context.Background(), planmodifier.StringRequest{
				StateValue: types.StringValue("a07abc12310546f1b9291ab301a7d75b"), PlanValue: tt.plan,
			}, &resp)
			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}

func TestClusterTrimmedStringKeepModifier(t *testing.T) {
	for _, tt := range []struct {
		name     string
		config   types.String
		expected types.String
	}{
		{name: "surrounding whitespace keeps the state value", config: types.StringValue("cert-1\n"), expected: types.StringValue("cert-1")},
		{name: "a different value is not rewritten", config: types.StringValue("cert-2"), expected: types.StringValue("cert-2")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := planmodifier.StringResponse{PlanValue: tt.config}
			trimmedStringKeep{}.PlanModifyString(context.Background(), planmodifier.StringRequest{
				StateValue: types.StringValue("cert-1"), ConfigValue: tt.config, PlanValue: tt.config,
			}, &resp)
			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}

func TestClusterCloudSubnetCIDRRequiresReplaceIf(t *testing.T) {
	r := &clusterResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	attr := schemaResp.Schema.Attributes["cloud_subnet_cidr"].(schema.StringAttribute)
	for _, tt := range []struct {
		name    string
		hasCIDR bool
		cidr    string
		replace bool
	}{
		{name: "an unset value after import does not replace", hasCIDR: false, cidr: "10.1.0.0/24", replace: false},
		{name: "a changed set value replaces", hasCIDR: true, cidr: "10.1.0.0/24", replace: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := clusterState()
			if !tt.hasCIDR {
				state.CloudSubnetCIDR = types.StringNull()
			}
			plan := state
			plan.CloudSubnetCIDR = types.StringValue(tt.cidr)
			req := planmodifier.StringRequest{
				Config: tfsdk.Config(clusterStateOf(t, r, plan)), Plan: tfsdk.Plan(clusterStateOf(t, r, plan)),
				State: clusterStateOf(t, r, state), StateValue: state.CloudSubnetCIDR, PlanValue: plan.CloudSubnetCIDR,
			}
			for _, m := range attr.PlanModifiers {
				resp := planmodifier.StringResponse{}
				m.PlanModifyString(context.Background(), req, &resp)
				assert.Equal(t, tt.replace, resp.RequiresReplace)
			}
		})
	}
}
