package mks

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testProviderConfig struct {
	client        *mksv2.ServiceClient
	clientErr     error
	project, pool string
	gotProject    []string
}

func (c *testProviderConfig) MKSV2Client(_ context.Context, projectID, _ string) (*mksv2.ServiceClient, error) {
	c.gotProject = append(c.gotProject, projectID)
	if c.clientErr != nil {
		return nil, c.clientErr
	}

	return c.client, nil
}

func (c *testProviderConfig) MKSV2Defaults() (string, string) { return c.project, c.pool }

func (f *fakeAPI) resource() *clusterResource {
	config := &testProviderConfig{client: f.client(), project: testProject, pool: "ru-1"}

	return &clusterResource{connection: connection{config: config}, pollInterval: time.Millisecond}
}

const (
	testProject        = apiProject
	testKubeVersion    = "1.31.8"
	testKubeVersionOld = "1.31.5"
	testKubeVersionUp  = "1.32.3"
)

func clusterState() clusterModel {
	kubeOptions := types.ObjectValueMust(kubeOptionsAttrTypes, map[string]attr.Value{
		"feature_gates":         types.SetValueMust(types.StringType, []attr.Value{types.StringValue("CSIMigration")}),
		"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("NodeRestriction")}),
		"audit_logs": types.ObjectValueMust(auditLogsAttrTypes, map[string]attr.Value{
			"enabled": types.BoolValue(false),
		}),
		"oidc": oidcObjectOf(nil),
	})

	return clusterModel{
		ID:                            types.StringValue(testClusterID),
		Name:                          types.StringValue("test-cluster"),
		Pool:                          types.StringValue("ru-1"),
		ProjectID:                     types.StringValue(testProject),
		KubeVersion:                   types.StringValue(testKubeVersion),
		ClusterType:                   types.StringValue("BASIC"),
		WorkersType:                   types.StringValue("CLOUD"),
		NetworkID:                     types.StringValue("fake-network"),
		SubnetID:                      types.StringValue("fake-subnet"),
		CloudSubnetCIDR:               types.StringValue("10.0.0.0/24"),
		MaintenanceWindowStart:        types.StringValue("06:00:00"),
		MaintenanceWindowEnd:          types.StringValue("09:00:00"),
		PrivateKubeAPI:                types.BoolValue(false),
		EnableAutorepair:              types.BoolValue(true),
		EnablePatchVersionAutoUpgrade: types.BoolValue(false),
		CNIType:                       types.StringValue("CALICO"),
		CNICiliumSettings:             types.ObjectNull(ciliumAttrTypes),
		KubernetesOptions:             kubeOptions,
		Status:                        types.StringValue("ACTIVE"),
		KubeAPIIP:                     types.StringValue("192.0.2.10"),
		Timeouts:                      timeoutsOf(2*time.Second, 2*time.Second, 2*time.Second),
	}
}

func timeoutsOf(create, update, del time.Duration) timeouts.Value {
	str := func(d time.Duration) types.String { return types.StringValue(d.String()) }

	return timeouts.Value{Object: types.ObjectValueMust(
		map[string]attr.Type{
			"create": types.StringType, "update": types.StringType, "delete": types.StringType,
		},
		map[string]attr.Value{"create": str(create), "update": str(update), "delete": str(del)},
	)}
}

func clusterPlan() clusterModel {
	m := clusterState()
	m.ID = types.StringUnknown()
	m.Status = types.StringUnknown()
	m.KubeAPIIP = types.StringUnknown()
	m.NetworkID = types.StringUnknown()
	m.SubnetID = types.StringUnknown()
	m.CloudSubnetCIDR = types.StringUnknown()
	m.MaintenanceWindowEnd = types.StringUnknown()

	return m
}

func clusterStateOf(t *testing.T, r *clusterResource, m clusterModel) tfsdk.State {
	t.Helper()
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	state := tfsdk.State{Schema: s.Schema}
	require.False(t, state.Set(context.Background(), &m).HasError())

	return state
}

func emptyState(t *testing.T, r *clusterResource) tfsdk.State {
	t.Helper()
	state := clusterStateOf(t, r, clusterState())
	state.Raw = tftypes.NewValue(state.Raw.Type(), nil)

	return state
}

func readClusterModel(t *testing.T, state tfsdk.State) clusterModel {
	t.Helper()
	var m clusterModel
	require.False(t, state.Get(context.Background(), &m).HasError())

	return m
}

func createCluster(t *testing.T, r *clusterResource, plan clusterModel) (tfsdk.State, diag.Diagnostics) {
	t.Helper()
	resp := resource.CreateResponse{State: emptyState(t, r)}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(clusterStateOf(t, r, plan))}, &resp)

	return resp.State, resp.Diagnostics
}

func readCluster(t *testing.T, r *clusterResource, prior clusterModel) (tfsdk.State, diag.Diagnostics) {
	t.Helper()
	state := clusterStateOf(t, r, prior)
	resp := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)

	return resp.State, resp.Diagnostics
}

func updateCluster(t *testing.T, r *clusterResource, prior, plan clusterModel) (clusterModel, diag.Diagnostics) {
	t.Helper()
	resp := resource.UpdateResponse{State: emptyState(t, r)}
	r.Update(context.Background(), resource.UpdateRequest{
		State: clusterStateOf(t, r, prior), Plan: tfsdk.Plan(clusterStateOf(t, r, plan)),
	}, &resp)
	if resp.State.Raw.IsNull() {
		return clusterModel{}, resp.Diagnostics
	}

	return readClusterModel(t, resp.State), resp.Diagnostics
}

func deleteCluster(t *testing.T, r *clusterResource, prior clusterModel) diag.Diagnostics {
	t.Helper()
	state := clusterStateOf(t, r, prior)
	resp := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)

	return resp.Diagnostics
}

func importCluster(t *testing.T, r *clusterResource) resource.ImportStateResponse {
	t.Helper()
	resp := resource.ImportStateResponse{State: emptyState(t, r)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: testClusterID}, &resp)

	return resp
}

func assertErrorContains(t *testing.T, diags diag.Diagnostics, parts ...string) {
	t.Helper()
	require.True(t, diags.HasError(), "want an error, got %v", diags)
	first := diags.Errors()[0]
	for _, part := range parts {
		assert.Contains(t, first.Summary()+": "+first.Detail(), part)
	}
}
