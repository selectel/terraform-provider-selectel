package mks

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/stretchr/testify/require"
)

// testProviderConfig implements ProviderConfig for the fake API server and
// records the last scope it was asked for.
type testProviderConfig struct {
	client        *mksv2.ServiceClient
	project, pool string
	gotProject    []string
	gotPool       []string
}

func (c *testProviderConfig) MKSV2Client(_ context.Context, projectID, pool string) (*mksv2.ServiceClient, error) {
	c.gotProject = append(c.gotProject, projectID)
	c.gotPool = append(c.gotPool, pool)

	return c.client, nil
}

func (c *testProviderConfig) MKSV2Defaults() (string, string) { return c.project, c.pool }

func (f *fakeAPI) resource() *clusterResource {
	config := &testProviderConfig{client: f.client(), project: testProject, pool: "ru-1"}

	return &clusterResource{connection: connection{config: config}}
}

const (
	// testProject is a project ID in the dashed form a plan carries.
	testProject = apiProject
	// testKubeVersions are what the fake kube versions serve: two patches of
	// 1.31 and one older minor.
	testKubeVersion    = "1.31.8"
	testKubeVersionOld = "1.31.5"
	testKubeVersionUp  = "1.32.3"
)

// clusterState is the settled state of a cluster: what fromAPI returns
// after the fake has created it.
func clusterState() clusterModel {
	kubeOptions := types.ObjectValueMust(kubeOptionsAttrTypes, map[string]attr.Value{
		"feature_gates":         types.SetValueMust(types.StringType, []attr.Value{types.StringValue("CSIMigration")}),
		"admission_controllers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("NodeRestriction")}),
		"audit_logs": types.ObjectValueMust(auditLogsAttrTypes, map[string]attr.Value{
			"enabled":     types.BoolValue(false),
			"secret_name": types.StringNull(),
		}),
		"oidc":                 oidcObject(),
		"x509_ca_certificates": types.StringNull(),
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
		Timeouts:                      shortTimeouts(),
	}
}

func oidcObject() types.Object {
	return types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
		"enabled":        types.BoolValue(false),
		"provider_name":  types.StringNull(),
		"issuer_url":     types.StringNull(),
		"client_id":      types.StringNull(),
		"username_claim": types.StringNull(),
		"groups_claim":   types.StringNull(),
		"ca_certs":       trimmedStringValue{StringValue: types.StringNull()},
	})
}

func shortTimeouts() timeouts.Value {
	return timeouts.Value{Object: types.ObjectValueMust(
		map[string]attr.Type{
			"create": types.StringType, "update": types.StringType, "delete": types.StringType,
		},
		map[string]attr.Value{
			"create": types.StringValue("2s"), "update": types.StringValue("2s"), "delete": types.StringValue("2s"),
		},
	)}
}

// clusterPlan is the plan of a new cluster: every computed value is unknown.
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
	planned := tfsdk.Plan(clusterStateOf(t, r, plan))
	resp := resource.UpdateResponse{State: tfsdk.State(planned)}
	r.Update(context.Background(), resource.UpdateRequest{State: clusterStateOf(t, r, prior), Plan: planned}, &resp)

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
	empty := clusterStateOf(t, r, clusterState())
	empty.Raw = tftypes.NewValue(empty.Raw.Type(), nil)
	resp := resource.ImportStateResponse{State: empty}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: testClusterID}, &resp)

	return resp
}

// firstError returns the detail of the first diagnostic error.
func firstError(t *testing.T, diags diag.Diagnostics) string {
	t.Helper()
	require.True(t, diags.HasError(), "want an error, got %v", diags)

	return diags.Errors()[0].Detail()
}

// errorText returns the summary and the detail of the first diagnostic
// error as one string.
func errorText(t *testing.T, diags diag.Diagnostics) string {
	t.Helper()
	require.True(t, diags.HasError(), "want an error, got %v", diags)
	first := diags.Errors()[0]

	return first.Summary() + ": " + first.Detail()
}
