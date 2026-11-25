package mks

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
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
		"audit_logs":            auditLogsObject(types.BoolValue(false)),
		"oidc":                  oidcObjectOf(nil),
	}
	maps.Copy(values, overrides)

	return types.ObjectValueMust(kubeOptionsAttrTypes, values)
}

func oidcObjectOf(fields map[string]attr.Value) types.Object {
	values := map[string]attr.Value{
		"enabled":        types.BoolValue(false),
		"provider_name":  types.StringNull(),
		"issuer_url":     types.StringNull(),
		"client_id":      types.StringNull(),
		"username_claim": types.StringNull(),
		"groups_claim":   types.StringNull(),
		"ca_certs":       trimmedStringValue{StringValue: types.StringNull()},
	}
	maps.Copy(values, fields)

	return types.ObjectValueMust(oidcAttrTypes, values)
}

func enabledOIDCObject() types.Object {
	return oidcObjectOf(map[string]attr.Value{
		"enabled":        types.BoolValue(true),
		"provider_name":  types.StringValue("keycloak"),
		"issuer_url":     types.StringValue("https://issuer.example.com"),
		"client_id":      types.StringValue("kubernetes"),
		"username_claim": types.StringValue("sub"),
		"groups_claim":   types.StringValue("groups"),
	})
}

func auditLogsObject(enabled types.Bool) types.Object {
	return types.ObjectValueMust(auditLogsAttrTypes, map[string]attr.Value{"enabled": enabled})
}

func ciliumObject(envoy, hubble types.Bool) types.Object {
	return types.ObjectValueMust(ciliumAttrTypes, map[string]attr.Value{
		"envoy_daemonset": envoy, "hubble_relay": hubble,
	})
}

func bodyFields(t *testing.T, body map[string]any, want map[string]any, absent ...string) {
	t.Helper()
	cluster, ok := body["cluster"].(map[string]any)
	if ok {
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

func kubeOptionsStateOf(t *testing.T, obj types.Object) clusterKubeOptionsModel {
	t.Helper()
	var options clusterKubeOptionsModel
	require.False(t, obj.As(context.Background(), &options, basetypes.ObjectAsOptions{}).HasError(), "%v", obj)

	return options
}

func fakeResource(t *testing.T) (*fakeAPI, *clusterResource) {
	t.Helper()
	fake := newFakeAPI(t)

	return fake, fake.resource()
}

func mustCreate(t *testing.T, r *clusterResource, plan clusterModel) clusterModel {
	t.Helper()
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	return readClusterModel(t, state)
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
	plan.CNICiliumSettings = ciliumObject(types.BoolValue(true), types.BoolValue(false))
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"audit_logs": auditLogsObject(types.BoolNull()),
		"oidc":       enabledOIDCObject(),
	})

	got := mustCreate(t, r, plan)

	body := fake.lastBody(routeCreateCluster)
	bodyFields(t, body, map[string]any{
		"network_type": "STANDARD", "basic": false, "cni_type": "CILIUM",
		"cni_cilium_settings": map[string]any{"envoy_daemonset": true, "hubble_relay": false},
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

	assert.Equal(t, ciliumObject(types.BoolValue(true), types.BoolValue(false)), got.CNICiliumSettings)
	assert.Equal(t, types.StringValue(testProject), got.ProjectID)
	assert.Equal(t, 1, fake.callCount(routeCreateCluster))
	assert.Equal(t, 1, fake.callCount(routeCluster))
	assert.Contains(t, fake.lastQuery(routeTasks), "limit=0")
}

func TestClusterProjectIDForms(t *testing.T) {
	keystone := keystoneProject
	for _, tt := range []struct {
		name       string
		planProj   string
		wantConfig string
	}{
		{name: "un-dashed configured", planProj: keystone, wantConfig: keystone},
		{name: "dashed configured", planProj: apiProject, wantConfig: apiProject},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, r := fakeResource(t)

			plan := clusterPlan()
			plan.ProjectID = types.StringValue(tt.planProj)
			got := mustCreate(t, r, plan)
			assert.Equal(t, types.StringValue(tt.wantConfig), got.ProjectID)
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
		autorepair  types.Bool
	}{
		{workersType: "CLOUD", networkType: "STANDARD", autorepair: types.BoolValue(true)},
		{workersType: "DEDICATED", networkType: "L3VPN", autorepair: types.BoolValue(false)},
	} {
		t.Run(tt.workersType, func(t *testing.T) {
			fake, r := fakeResource(t)

			plan := clusterPlan()
			plan.WorkersType = types.StringValue(tt.workersType)
			plan.EnableAutorepair = tt.autorepair
			state, diags := createCluster(t, r, plan)
			require.False(t, diags.HasError(), "%v", diags)

			bodyFields(t, fake.lastBody(routeCreateCluster), map[string]any{"network_type": tt.networkType})

			got := readClusterModel(t, state)
			assert.Equal(t, types.StringValue(tt.workersType), got.WorkersType)
			assert.Equal(t, tt.autorepair, got.EnableAutorepair)
		})
	}
}

func TestClusterCreateErrors(t *testing.T) {
	for _, tt := range []struct {
		name    string
		setup   func(fake *fakeAPI)
		wantErr string
	}{
		{
			name:  "a task that ends in ERROR reports the task chain",
			setup: func(fake *fakeAPI) { fake.failTasks("CREATE_CLUSTER", true) },
			wantErr: "task CREATE_CLUSTER task-1 of cluster " + testClusterID +
				" ended in ERROR: fake failure (code 42): fake details of CREATE_CLUSTER",
		},
		{
			name: "a 400 with the API error",
			setup: func(fake *fakeAPI) {
				fake.failWithBody(routeCreateCluster, http.StatusBadRequest, "application/json", `{"error":{"message":"invalid kube_version: 1.31.8"}}`)
			},
			wantErr: "error creating cluster: 400 Bad Request: invalid kube_version: 1.31.8",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t)
			tt.setup(fake)
			r := fake.resource()

			_, diags := createCluster(t, r, clusterPlan())
			require.True(t, diags.HasError())
			assert.Contains(t, diags.Errors()[0].Detail(), tt.wantErr)
		})
	}
}

func TestClusterPatchBodies(t *testing.T) {
	t.Run("kubernetes_options go whole, other fields only when changed", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()

		prior := mustCreate(t, r, clusterPlan())

		optionsPlan := prior
		optionsPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
			"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("NodeRestriction"), types.StringValue("PodNodeSelector"),
			}),
			"audit_logs": auditLogsObject(types.BoolValue(true)),
			"oidc":       enabledOIDCObject(),
		})
		_, diags := updateCluster(t, r, prior, optionsPlan)
		require.False(t, diags.HasError(), "%v", diags)
		cluster, _ := fake.lastBody(routePatchCluster)["cluster"].(map[string]any)
		require.Len(t, cluster, 1)
		options, _ := cluster["kubernetes_options"].(map[string]any)
		assert.Equal(t, "[TopologyAwareHints]", fmt.Sprint(options["feature_gates"]))
		audit, _ := options["audit_logs"].(map[string]any)
		assert.Equal(t, true, audit["enabled"])
		oidc, _ := options["oidc"].(map[string]any)
		assert.Equal(t, true, oidc["enabled"])
		assert.Greater(t, fake.callCount(routeTask), 0, "the provider must poll the UPGRADE_MASTERS_CONFIG task")

		windowPlan := prior
		windowPlan.MaintenanceWindowStart = types.StringValue("07:00:00")
		got, diags := updateCluster(t, r, prior, windowPlan)
		require.False(t, diags.HasError(), "%v", diags)
		cluster, _ = fake.lastBody(routePatchCluster)["cluster"].(map[string]any)
		require.Len(t, cluster, 1)
		assert.Equal(t, "07:00:00", cluster["maintenance_window_start"])

		autorepairPlan := got
		autorepairPlan.EnableAutorepair = types.BoolValue(false)
		_, diags = updateCluster(t, r, got, autorepairPlan)
		require.False(t, diags.HasError(), "%v", diags)
		cluster, _ = fake.lastBody(routePatchCluster)["cluster"].(map[string]any)
		require.Len(t, cluster, 1)
		assert.Equal(t, false, cluster["enable_autorepair"])
	})

	t.Run("enable_patch_version_auto_upgrade round-trips false, true, false", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()

		plan := clusterPlan()
		plan.EnablePatchVersionAutoUpgrade = types.BoolValue(false)
		state, diags := createCluster(t, r, plan)
		require.False(t, diags.HasError(), "%v", diags)
		bodyFields(t, fake.lastBody(routeCreateCluster), map[string]any{"enable_patch_version_auto_upgrade": false})

		got := readClusterModel(t, state)
		for _, enabled := range []bool{true, false} {
			next := got
			next.EnablePatchVersionAutoUpgrade = types.BoolValue(enabled)
			got, diags = updateCluster(t, r, got, next)
			require.False(t, diags.HasError(), "%v", diags)
			bodyFields(t, fake.lastBody(routePatchCluster), map[string]any{
				"enable_patch_version_auto_upgrade": enabled,
			})
		}
	})
}

func TestClusterKubeVersionUpgrade(t *testing.T) {
	fake, r := fakeResource(t)
	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new(testKubeVersionOld)},
		mksclient.KubeVersionInfo{Version: new(testKubeVersion)},
		mksclient.KubeVersionInfo{Version: new(testKubeVersionUp)},
	)

	plan := clusterPlan()
	plan.KubeVersion = types.StringValue(testKubeVersionOld)

	prior := mustCreate(t, r, plan)
	assert.Equal(t, 0, fake.callCount(routeUpgradePatch))
	assert.Equal(t, 0, fake.callCount(routeUpgradeMinor))

	badPlan := prior
	badPlan.KubeVersion = types.StringValue("1.31.7")
	_, diags := updateCluster(t, r, prior, badPlan)
	assertErrorContains(t, diags, fmt.Sprintf(
		"current version %s can't be upgraded to version 1.31.7, the latest available patch version is: %s",
		testKubeVersionOld, testKubeVersion))

	okPlan := prior
	okPlan.KubeVersion = types.StringValue(testKubeVersion)
	got, diags := updateCluster(t, r, prior, okPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersion), got.KubeVersion)
	assert.Equal(t, 1, fake.callCount(routeUpgradePatch))
	assert.Equal(t, 0, fake.callCount(routeUpgradeMinor))

	up := got
	up.KubeVersion = types.StringValue(testKubeVersionUp)
	got2, diags := updateCluster(t, r, got, up)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue(testKubeVersionUp), got2.KubeVersion)

	bad := got2
	bad.KubeVersion = types.StringValue("1.34.1")
	_, diags = updateCluster(t, r, got2, bad)
	assertErrorContains(t, diags, "kubernetes versions must be upgraded one minor version at a time")
	_, diags = updateCluster(t, r, got2, withKubeVersion(got2, "2.33.1"))
	assertErrorContains(t, diags, "kubernetes versions must be upgraded one minor version at a time")
	assert.Equal(t, 1, fake.callCount(routeUpgradeMinor))

	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.KubeVersion = "1.32.5" })
	behind := withKubeVersion(got2, testKubeVersionOld)
	_, diags = updateCluster(t, r, behind, withKubeVersion(behind, testKubeVersion))
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, 1, fake.callCount(routeUpgradePatch))
	assert.Equal(t, 1, fake.callCount(routeUpgradeMinor))
}

func TestClusterMinorUpgradeAcceptsAnyPatchOfTheNextMinor(t *testing.T) {
	fake, r := fakeResource(t)
	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new(testKubeVersion)},
		mksclient.KubeVersionInfo{Version: new("1.32.1")},
		mksclient.KubeVersionInfo{Version: new(testKubeVersionUp)},
	)
	prior := mustCreate(t, r, clusterPlan())

	got, diags := updateCluster(t, r, prior, withKubeVersion(prior, "1.32.1"))

	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("1.32.1"), got.KubeVersion)
	assert.Equal(t, 1, fake.callCount(routeUpgradeMinor))
	assert.Equal(t, 0, fake.callCount(routeUpgradePatch))
}

func TestClusterMinorUpgradeStoresTheRemainingKubernetesOptions(t *testing.T) {
	fake, r := fakeResource(t)
	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new(testKubeVersion)},
		mksclient.KubeVersionInfo{Version: new(testKubeVersionUp)},
	)
	prior := mustCreate(t, r, clusterPlan())
	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) {
		c.KubernetesOptions.FeatureGates = []string{}
	})

	plan := withKubeVersion(prior, testKubeVersionUp)
	plan.KubernetesOptions = types.ObjectUnknown(kubeOptionsAttrTypes)
	got, diags := updateCluster(t, r, prior, plan)

	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, 1, fake.callCount(routeUpgradeMinor))
	assert.Equal(t, 0, fake.callCount(routePatchCluster))
	options := kubeOptionsStateOf(t, got.KubernetesOptions)
	assert.Empty(t, options.FeatureGates.Elements())
	assert.Len(t, options.AdmissionControllers.Elements(), 1)
}

func TestClusterMinorUpgradeWithoutANextMinorFails(t *testing.T) {
	fake, r := fakeResource(t)
	fake.seedKubeVersions(mksclient.KubeVersionInfo{Version: new(testKubeVersion)})
	prior := mustCreate(t, r, clusterPlan())

	_, diags := updateCluster(t, r, prior, withKubeVersion(prior, testKubeVersionUp))

	assertErrorContains(t, diags, "Error upgrading cluster", "the cluster is already on the latest available minor version: 1.31")
	assert.Equal(t, 1, fake.callCount(routeUpgradeMinor))
}

func TestClusterReadStoresTheAPIKubeVersion(t *testing.T) {
	fake, r := fakeResource(t)
	prior := mustCreate(t, r, clusterPlan())

	for _, actual := range []string{"1.31.5", "1.31.9", "1.32.1"} {
		fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.KubeVersion = actual })
		state, diags := readCluster(t, r, prior)
		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, types.StringValue(actual), readClusterModel(t, state).KubeVersion)
	}
}

func TestClusterMinorKubeVersion(t *testing.T) {
	fake, r := fakeResource(t)
	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new(testKubeVersionOld)},
		mksclient.KubeVersionInfo{Version: new(testKubeVersionUp)},
	)
	readAt := func(prior clusterModel, actual string) types.String {
		fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.KubeVersion = actual })
		state, diags := readCluster(t, r, prior)
		require.False(t, diags.HasError(), "%v", diags)

		return readClusterModel(t, state).KubeVersion
	}

	prior := mustCreate(t, r, withKubeVersion(clusterPlan(), "1.31"))
	bodyFields(t, fake.lastBody(routeCreateCluster), map[string]any{"kube_version": "1.31"})
	assert.Equal(t, types.StringValue("1.31"), prior.KubeVersion)
	assert.Equal(t, types.StringValue("1.31"), readAt(prior, testKubeVersionOld))
	assert.Equal(t, types.StringValue("1.31"), readAt(prior, testKubeVersion))

	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new("1.31.9")},
		mksclient.KubeVersionInfo{Version: new(testKubeVersionUp)},
	)
	patched, diags := updateCluster(t, r, prior, withKubeVersion(prior, "1.31.9"))
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("1.31.9"), patched.KubeVersion)
	assert.Equal(t, 1, fake.callCount(routeUpgradePatch))

	upgraded, diags := updateCluster(t, r, prior, withKubeVersion(prior, "1.32"))
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("1.32"), upgraded.KubeVersion)
	assert.Equal(t, 1, fake.callCount(routeUpgradeMinor))
	assert.Equal(t, types.StringValue("1.33.1"), readAt(upgraded, "1.33.1"))
}

func TestClusterUpdateUpgradesBeforeThePatch(t *testing.T) {
	fake, r := fakeResource(t)
	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new(testKubeVersionOld)},
		mksclient.KubeVersionInfo{Version: new(testKubeVersion)},
	)
	plan := withKubeVersion(clusterPlan(), testKubeVersionOld)
	prior := mustCreate(t, r, plan)

	next := withKubeVersion(prior, testKubeVersion)
	next.KubernetesOptions = kubeOptionsObject(nil)
	_, diags := updateCluster(t, r, prior, next)

	require.False(t, diags.HasError(), "%v", diags)
	taskTypes := make([]string, 0, len(fake.tasks))
	for _, tk := range fake.tasks {
		taskTypes = append(taskTypes, tk.task.Type)
	}
	assert.Equal(t, []string{"CREATE_CLUSTER", "UPGRADE_PATCH_VERSION", "UPGRADE_MASTERS_CONFIG"}, taskTypes)
}

func TestClusterImport(t *testing.T) {
	_, r := fakeResource(t)
	config := r.config.(*testProviderConfig)
	config.pool = "ru-2"

	resp := importCluster(t, r)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	got := readClusterModel(t, resp.State)
	assert.Equal(t, types.StringValue(testClusterID), got.ID)
	assert.Equal(t, types.StringValue(keystoneProject), got.ProjectID)
	assert.Equal(t, types.StringValue("ru-2"), got.Pool)

	plan := clusterPlan()
	plan.ProjectID = types.StringNull()
	_, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	importedResp := importCluster(t, r)
	require.False(t, importedResp.Diagnostics.HasError(), "%v", importedResp.Diagnostics)
	imported := readClusterModel(t, importedResp.State)
	assert.Equal(t, types.StringValue(keystoneProject), imported.ProjectID)

	state2, diags := readCluster(t, r, imported)
	require.False(t, diags.HasError(), "%v", diags)
	kube := readClusterModel(t, state2)
	assert.Equal(t, types.StringValue(testClusterID), kube.ID)
	assert.Equal(t, types.StringValue(keystoneProject), kube.ProjectID)
	assert.True(t, kube.CloudSubnetCIDR.IsNull())
	assert.Equal(t, types.StringValue(testKubeVersion), kube.KubeVersion)
}

func TestClusterImportNeedsProviderConfig(t *testing.T) {
	for _, tt := range []struct {
		name   string
		config *testProviderConfig
		want   string
	}{
		{name: "no configuration", config: nil, want: "INFRA_PROJECT_ID must be set for the resource import"},
		{name: "no project", config: &testProviderConfig{pool: "ru-1"}, want: "INFRA_PROJECT_ID must be set for the resource import"},
		{name: "no pool", config: &testProviderConfig{project: testProject}, want: "INFRA_REGION must be set for the resource import"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var config ProviderConfig
			if tt.config != nil {
				config = tt.config
			}
			r := &clusterResource{connection: connection{config: config}}
			resp := importCluster(t, r)

			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tt.want)
		})
	}
}

func TestClusterDeleteTaskError(t *testing.T) {
	fake, r := fakeResource(t)

	plan := clusterPlan()
	plan.KubeVersion = types.StringValue(testKubeVersion)
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	fake.failTasks("DELETE_CLUSTER", true)
	diags = deleteCluster(t, r, readClusterModel(t, state))
	assertErrorContains(t, diags, "task DELETE_CLUSTER task-", "ended in ERROR")

	fake.failTasks("DELETE_CLUSTER", false)
	diags = deleteCluster(t, r, readClusterModel(t, state))
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, 0, fake.clusterCount())
}

func TestClusterWaits(t *testing.T) {
	patchPlan := func(got clusterModel) clusterModel {
		got.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
			"feature_gates": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("NewFeatureGate")}),
		})

		return got
	}
	for _, tt := range []struct {
		name  string
		stage func(t *testing.T, fake *fakeAPI, r *clusterResource)
	}{
		{
			name: "a stuck create wait times out and still saves the cluster",
			stage: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				fake.stickTasks("CREATE_CLUSTER", true)
				plan := clusterPlan()
				plan.Timeouts = timeoutsOf(time.Second, 2*time.Second, 2*time.Second)
				savedState, diags := createCluster(t, r, plan)
				assertErrorContains(t, diags, "Error waiting for the cluster to become ready", "last state: task CREATE_CLUSTER", "context deadline exceeded")
				assert.Equal(t, 1, fake.callCount(routeCreateCluster))
				assert.Equal(t, types.StringValue(testClusterID), readClusterModel(t, savedState).ID)

				fake.stickTasks("CREATE_CLUSTER", false)
				fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "ACTIVE" })
				assert.False(t, deleteCluster(t, r, readClusterModel(t, savedState)).HasError())
			},
		},
		{
			name: "a stuck create wait with a failing read reports both",
			stage: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				fake.stickTasks("CREATE_CLUSTER", true)
				fake.fail(routeCluster, http.StatusInternalServerError)
				plan := clusterPlan()
				plan.Timeouts = timeoutsOf(time.Second, 2*time.Second, 2*time.Second)
				savedState, diags := createCluster(t, r, plan)
				assertErrorContains(t, diags, "Error reading cluster", "context deadline exceeded")
				saved := readClusterModel(t, savedState)
				assert.Equal(t, types.StringValue(testClusterID), saved.ID)
				assert.Equal(t, plan.Name, saved.Name)
			},
		},
		{
			name: "a stuck update wait times out and the next refresh sees the change",
			stage: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				fake.stickTasks("UPGRADE_MASTERS_CONFIG", true)
				plan := clusterPlan()
				plan.Timeouts = timeoutsOf(2*time.Second, time.Second, 2*time.Second)
				got := mustCreate(t, r, plan)

				_, diags := updateCluster(t, r, got, patchPlan(got))
				assertErrorContains(t, diags, "Error updating cluster")
				assertErrorContains(t, diags, "context deadline exceeded")

				fake.stickTasks("UPGRADE_MASTERS_CONFIG", false)
				fake.settleAfter(testClusterID, 1, "ACTIVE")
				refreshedState, diags := readCluster(t, r, got)
				require.False(t, diags.HasError(), "%v", diags)
				refreshed := readClusterModel(t, refreshedState)
				assert.Contains(t, refreshed.KubernetesOptions.String(), "NewFeatureGate")
			},
		},
		{
			name: "a stuck delete wait times out",
			stage: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				plan := clusterPlan()
				plan.Timeouts = timeoutsOf(2*time.Second, 2*time.Second, time.Second)
				got := mustCreate(t, r, plan)

				fake.stickTasks("DELETE_CLUSTER", true)
				diags := deleteCluster(t, r, got)
				assertErrorContains(t, diags, "Error deleting cluster")
				assertErrorContains(t, diags, "context deadline exceeded")
				assert.Equal(t, 1, fake.callCount(routeDeleteCluster))
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake, r := fakeResource(t)
			tt.stage(t, fake, r)
		})
	}
}

func TestClusterStatusOnUpdate(t *testing.T) {
	fake, r := fakeResource(t)

	plan := clusterPlan()
	plan.CNIType = types.StringValue("CILIUM")
	plan.CNICiliumSettings = ciliumObject(types.BoolValue(true), types.BoolValue(false))
	prior := mustCreate(t, r, plan)

	ciliumPlan := prior
	ciliumPlan.CNICiliumSettings = ciliumObject(types.BoolValue(true), types.BoolValue(true))
	got, diags := updateCluster(t, r, prior, ciliumPlan)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, 1, fake.callCount(routePatchCluster))
	assert.Equal(t, types.StringValue("ACTIVE"), got.Status)
	ciliumPatch, _ := fake.lastBody(routePatchCluster)["cluster"].(map[string]any)
	assert.Equal(t, map[string]any{"envoy_daemonset": true, "hubble_relay": true}, ciliumPatch["cni_cilium_settings"])
}

func TestClusterPatchConflictFailsWithTheAPIText(t *testing.T) {
	fake, r := fakeResource(t)
	prior := mustCreate(t, r, clusterPlan())

	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "PENDING_UPDATE" })
	_, diags := updateCluster(t, r, prior, withAutorepair(prior, false))

	assertErrorContains(t, diags, "Error updating cluster", "cluster or one of nodegroups has invalid status to perform this operation")
	assert.Equal(t, 1, fake.callCount(routePatchCluster))
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
	prior := mustCreate(t, r, plan)

	fake.failTasks("UPGRADE_PATCH_VERSION", true)
	upPlan := prior
	upPlan.KubeVersion = types.StringValue(testKubeVersion)
	_, diags := updateCluster(t, r, prior, upPlan)
	assertErrorContains(t, diags, "task UPGRADE_PATCH_VERSION task-", "ended in ERROR")

	state, diags := readCluster(t, r, prior)
	require.False(t, diags.HasError(), "%v", diags)
	got := readClusterModel(t, state)
	assert.Equal(t, types.StringValue(testKubeVersion), got.KubeVersion)
	assert.Equal(t, types.StringValue("ERROR"), got.Status)

	assert.Equal(t, 1, fake.callCount(routeCreateCluster))
	assert.Equal(t, 1, fake.callCount(routeUpgradePatch))

	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "ACTIVE" })
	fake.failTasks("UPGRADE_MASTERS_CONFIG", true)
	mastersPlan := got
	mastersPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("NodeRestriction"), types.StringValue("PodNodeSelector"),
		}),
	})
	_, diags = updateCluster(t, r, got, mastersPlan)
	assertErrorContains(t, diags, "task UPGRADE_MASTERS_CONFIG task-", "ended in ERROR")

	fake.failTasks("UPGRADE_MASTERS_CONFIG", false)
	state2, diags := readCluster(t, r, got)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Contains(t, readClusterModel(t, state2).KubernetesOptions.String(), "PodNodeSelector")
}

func TestClusterCreateUnknownFieldsLeftToAPI(t *testing.T) {
	fake, r := fakeResource(t)

	plan := clusterPlan()
	plan.EnablePatchVersionAutoUpgrade = types.BoolUnknown()
	plan.CNIType = types.StringUnknown()
	plan.CNICiliumSettings = types.ObjectUnknown(ciliumAttrTypes)
	plan.KubernetesOptions = types.ObjectUnknown(kubeOptionsAttrTypes)
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	bodyFields(t, fake.lastBody(routeCreateCluster), nil,
		"enable_patch_version_auto_upgrade", "cni_type", "cni_cilium_settings", "kubernetes_options")
	got := readClusterModel(t, state)
	assert.Equal(t, types.BoolValue(false), got.EnablePatchVersionAutoUpgrade)
	assert.Equal(t, types.StringValue("CALICO"), got.CNIType)
	assert.True(t, got.CNICiliumSettings.IsNull())
}

func TestClusterOIDCCACertsWhitespace(t *testing.T) {
	fake, r := fakeResource(t)

	plan := clusterPlan()
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": oidcObjectOf(map[string]attr.Value{
			"enabled": types.BoolValue(true), "provider_name": types.StringValue("keycloak"),
			"issuer_url": types.StringValue("https://issuer.example.com"),
			"client_id":  types.StringValue("kubernetes"), "username_claim": types.StringValue("sub"),
			"groups_claim": types.StringValue("groups"), "ca_certs": trimmedString("cert-1\n"),
		}),
	})
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)

	body := oidcOf(fake.lastBody(routeCreateCluster))
	assert.Equal(t, "cert-1\n", body["ca_certs"])
	var oidc clusterOIDCModel
	require.False(t, kubeOptionsStateOf(t, readClusterModel(t, state).KubernetesOptions).OIDC.As(context.Background(), &oidc, basetypes.ObjectAsOptions{}).HasError())
	same, diags := oidc.CACerts.StringSemanticEquals(context.Background(), trimmedString(" cert-1 "))
	require.False(t, diags.HasError())
	assert.True(t, same)
}

func TestClusterOIDCDisable(t *testing.T) {
	fake, r := fakeResource(t)

	plan := clusterPlan()
	oidcFull := oidcObjectOf(map[string]attr.Value{
		"enabled": types.BoolValue(true), "provider_name": types.StringValue("keycloak"),
		"issuer_url": types.StringValue("https://issuer.example.com"),
		"client_id":  types.StringValue("kubernetes"), "username_claim": types.StringValue("sub"),
		"groups_claim": types.StringValue("groups"), "ca_certs": trimmedString("cert"),
	})
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidcFull})
	prior := mustCreate(t, r, plan)

	newPlan := prior
	newPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidcObjectOf(nil)})
	got, diags := updateCluster(t, r, prior, newPlan)
	require.False(t, diags.HasError(), "%v", diags)
	oidc := oidcOf(fake.lastBody(routePatchCluster))
	assert.Equal(t, map[string]any{
		"enabled": false, "provider_name": "", "issuer_url": "", "client_id": "", "username_claim": "", "groups_claim": "",
	}, oidc)
	assert.Contains(t, got.KubernetesOptions.String(), `"enabled":false`)

	reEnable := prior
	_, diags = updateCluster(t, r, got, reEnable)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, map[string]any{
		"enabled": true, "provider_name": "keycloak", "issuer_url": "https://issuer.example.com",
		"client_id": "kubernetes", "username_claim": "sub", "groups_claim": "groups", "ca_certs": "cert",
	}, oidcOf(fake.lastBody(routePatchCluster)))
}

func withKubeVersion(m clusterModel, version string) clusterModel {
	plan := m
	plan.KubeVersion = types.StringValue(version)

	return plan
}

func withAutorepair(m clusterModel, enabled bool) clusterModel {
	plan := m
	plan.EnableAutorepair = types.BoolValue(enabled)

	return plan
}

func TestClusterCRUDFailsBeforeTheAPI(t *testing.T) {
	for _, tt := range []struct {
		name string
		run  func(t *testing.T, fake *fakeAPI, r *clusterResource)
	}{
		{
			name: "missing project fails every CRUD path with one error and no API call",
			run: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				r.config = &testProviderConfig{client: fake.client(), project: "", pool: ""}

				plan := clusterPlan()
				plan.ProjectID = types.StringNull()
				prior := clusterState()
				prior.ProjectID = types.StringNull()
				for _, call := range []func() diag.Diagnostics{
					func() diag.Diagnostics { _, d := createCluster(t, r, plan); return d },
					func() diag.Diagnostics { _, d := updateCluster(t, r, prior, plan); return d },
					func() diag.Diagnostics { _, d := readCluster(t, r, prior); return d },
					func() diag.Diagnostics { return deleteCluster(t, r, prior) },
				} {
					d := call()
					require.True(t, d.HasError())
					assert.Equal(t, "Missing project ID", d.Errors()[0].Summary())
					assert.Len(t, d.Errors(), 1)
				}

				assert.Equal(t, 0, fake.callCount(routeCluster))
				assert.Equal(t, 0, fake.callCount(routeDeleteCluster))
				assert.Equal(t, 0, fake.clusterCount())
			},
		},
		{
			name: "a failing client fails every CRUD path",
			run: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				cfg := r.config.(*testProviderConfig)
				cfg.clientErr = errors.New("no such auth token")

				renamed := clusterState()
				renamed.Name = types.StringValue("renamed")
				for _, call := range []func() diag.Diagnostics{
					func() diag.Diagnostics { _, d := createCluster(t, r, clusterPlan()); return d },
					func() diag.Diagnostics { _, d := readCluster(t, r, clusterState()); return d },
					func() diag.Diagnostics { _, d := updateCluster(t, r, clusterState(), renamed); return d },
					func() diag.Diagnostics { return deleteCluster(t, r, clusterState()) },
				} {
					assertErrorContains(t, call(), "Error initializing MKS client", "connect to Managed Kubernetes")
				}
				assert.Equal(t, 0, fake.callCount(routeCluster))
			},
		},
		{
			name: "a 500 on read carries the cluster id",
			run: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				fake.fail(routeCluster, http.StatusInternalServerError)

				_, diags := readCluster(t, r, clusterState())
				assertErrorContains(t, diags, "Error reading cluster", "'"+testClusterID+"'")
			},
		},
		{
			name: "a 404 is a removal in Read only",
			run: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				got := mustCreate(t, r, clusterPlan())
				fake.removeCluster(testClusterID)

				state2, diags := readCluster(t, r, got)
				require.False(t, diags.HasError(), "%v", diags)
				assert.True(t, state2.Raw.IsNull())
			},
		},
		{
			name: "the upgrade of a gone cluster fails, the delete stays clean",
			run: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				prior := mustCreate(t, r, clusterPlan())
				fake.removeCluster(testClusterID)

				_, diags := updateCluster(t, r, prior, withKubeVersion(prior, testKubeVersionOld))
				assertErrorContains(t, diags, "Error upgrading cluster", "'"+testClusterID+"'")

				diags = deleteCluster(t, r, prior)
				assert.False(t, diags.HasError(), "%v", diags)
			},
		},
		{
			name: "a read-back error after a PATCH fails the update",
			run: func(t *testing.T, fake *fakeAPI, r *clusterResource) {
				prior := mustCreate(t, r, clusterPlan())

				fake.settleAfter(testClusterID, 2, "")
				_, diags := updateCluster(t, r, prior, withAutorepair(prior, false))
				assertErrorContains(t, diags, "Error reading cluster")
				assert.Equal(t, 1, fake.callCount(routePatchCluster))
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake, r := fakeResource(t)
			tt.run(t, fake, r)
		})
	}
}

func TestClusterCloudSubnetCIDRComesFromTheAPI(t *testing.T) {
	fake, r := fakeResource(t)

	plan := clusterPlan()
	plan.WorkersType = types.StringValue("DEDICATED")
	plan.EnableAutorepair = types.BoolValue(false)
	created := mustCreate(t, r, plan)
	assert.Equal(t, types.StringValue("10.20.0.0/24"), created.CloudSubnetCIDR)

	imported := readClusterModel(t, importCluster(t, r).State)
	state, diags := readCluster(t, r, imported)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, types.StringValue("10.20.0.0/24"), readClusterModel(t, state).CloudSubnetCIDR)

	plan.CloudSubnetCIDR = types.StringValue("10.30.0.0/24")
	got := mustCreate(t, r, plan)
	bodyFields(t, fake.lastBody(routeCreateCluster), map[string]any{"cloud_subnet_cidr": "10.30.0.0/24"})
	assert.Equal(t, types.StringValue("10.30.0.0/24"), got.CloudSubnetCIDR)
}

func TestClusterSchemaDefaults(t *testing.T) {
	var resp resource.SchemaResponse
	(&clusterResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	cilium := resp.Schema.Attributes["cni_cilium_settings"].(schema.SingleNestedAttribute)
	options := resp.Schema.Attributes["kubernetes_options"].(schema.SingleNestedAttribute)

	var hubble defaults.BoolResponse
	cilium.Attributes["hubble_relay"].(schema.BoolAttribute).Default.DefaultBool(context.Background(), defaults.BoolRequest{}, &hubble)
	assert.Equal(t, types.BoolValue(true), hubble.PlanValue)

	for _, name := range []string{"feature_gates", "admission_controllers"} {
		var set defaults.SetResponse
		options.Attributes[name].(schema.SetAttribute).Default.DefaultSet(context.Background(), defaults.SetRequest{}, &set)
		assert.Equal(t, types.SetValueMust(types.StringType, nil), set.PlanValue, name)
	}
}

func TestClusterRemovedFeatureGatesSendAnEmptyList(t *testing.T) {
	fake, r := fakeResource(t)
	prior := mustCreate(t, r, clusterPlan())

	next := prior
	next.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"feature_gates": types.SetValueMust(types.StringType, nil),
	})
	got, diags := updateCluster(t, r, prior, next)
	require.False(t, diags.HasError(), "%v", diags)

	cluster, _ := fake.lastBody(routePatchCluster)["cluster"].(map[string]any)
	options, _ := cluster["kubernetes_options"].(map[string]any)
	assert.Equal(t, []any{}, options["feature_gates"])
	gates := kubeOptionsStateOf(t, got.KubernetesOptions).FeatureGates
	assert.False(t, gates.IsNull())
	assert.Empty(t, gates.Elements())
}

func TestClusterWaitsForActiveAfterTheTasks(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		fake, r := fakeResource(t)
		fake.settleAfter(testClusterID, 1, "ERROR")

		state, diags := createCluster(t, r, clusterPlan())

		assertErrorContains(t, diags, "Error waiting for the cluster to become ready", "cluster "+testClusterID+" is in ERROR")
		assert.Equal(t, types.StringValue(testClusterID), readClusterModel(t, state).ID)
	})

	t.Run("update", func(t *testing.T) {
		fake, r := fakeResource(t)
		prior := mustCreate(t, r, clusterPlan())
		fake.settleAfter(testClusterID, 1, "ERROR")

		_, diags := updateCluster(t, r, prior, withAutorepair(prior, false))

		assertErrorContains(t, diags, "Error updating cluster", "cluster "+testClusterID+" is in ERROR")
	})
}
