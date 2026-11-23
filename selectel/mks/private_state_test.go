package mks

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type privateStateTestProvider struct {
	config *testProviderConfig
}

func (p privateStateTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "selectel"
}

func (p privateStateTestProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {
}

func (p privateStateTestProvider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	resp.ResourceData = p.config
}

func (p privateStateTestProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewClusterResource}
}

func configureClusterProvider(t *testing.T, config *testProviderConfig) tfprotov6.ProviderServer {
	t.Helper()

	server := providerserver.NewProtocol6(privateStateTestProvider{config: config})()
	resp, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{})
	require.NoError(t, err)
	require.Empty(t, formatDiagnostics(resp.Diagnostics), "%s", formatDiagnostics(resp.Diagnostics))

	return server
}

func (p privateStateTestProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}

func TestClusterImportStateSetsImportedKey(t *testing.T) {
	fake := newFakeAPI(t)
	fake.seedCluster(testClusterID)
	ctx := context.Background()
	server := configureClusterProvider(t, &testProviderConfig{client: fake.client(), project: testProject, pool: "ru-1"})

	resp, err := server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{
		TypeName: "selectel_mks_cluster_v2", ID: testClusterID,
	})
	require.NoError(t, err)
	require.Empty(t, formatDiagnostics(resp.Diagnostics), "%s", formatDiagnostics(resp.Diagnostics))
	require.Len(t, resp.ImportedResources, 1)

	imported := resp.ImportedResources[0]
	var privateKeys map[string][]byte
	require.NoError(t, json.Unmarshal(imported.Private, &privateKeys))
	assert.Equal(t, "true", string(privateKeys[importedKey]))
}

func TestClusterCloudSubnetCIDRImportedPlanKeepsTheCluster(t *testing.T) {
	fake := newFakeAPI(t)
	fake.seedCluster(testClusterID)
	ctx := context.Background()
	server := configureClusterProvider(t, &testProviderConfig{client: fake.client(), project: testProject, pool: "ru-1"})
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	valueType := schemas.ResourceSchemas["selectel_mks_cluster_v2"].ValueType()

	importResp, err := server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{
		TypeName: "selectel_mks_cluster_v2", ID: testClusterID,
	})
	require.NoError(t, err)
	require.Empty(t, formatDiagnostics(importResp.Diagnostics), "%s", formatDiagnostics(importResp.Diagnostics))

	imported, err := importResp.ImportedResources[0].State.Unmarshal(valueType)
	require.NoError(t, err)
	var priorAttrs map[string]tftypes.Value
	require.NoError(t, imported.As(&priorAttrs))
	require.True(t, priorAttrs["cloud_subnet_cidr"].IsNull())

	prior := clusterState()
	prior.ID = types.StringValue(testClusterID)
	prior.CloudSubnetCIDR = types.StringNull()

	config := prior
	config.CloudSubnetCIDR = types.StringValue("10.1.0.0/24")
	proposed := prior
	proposed.CloudSubnetCIDR = types.StringValue("10.1.0.0/24")

	encode := func(m clusterModel) *tfprotov6.DynamicValue {
		value, err := tfprotov6.NewDynamicValue(valueType, clusterStateOf(t, fake.resource(), m).Raw)
		require.NoError(t, err)

		return &value
	}

	planResp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName:         "selectel_mks_cluster_v2",
		PriorState:       encode(prior),
		ProposedNewState: encode(proposed),
		Config:           encode(config),
		PriorPrivate:     importResp.ImportedResources[0].Private,
	})
	require.NoError(t, err)
	require.Empty(t, formatDiagnostics(planResp.Diagnostics), "%s", formatDiagnostics(planResp.Diagnostics))
	assert.Empty(t, planResp.RequiresReplace)

	noImportResp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName:         "selectel_mks_cluster_v2",
		PriorState:       encode(prior),
		ProposedNewState: encode(proposed),
		Config:           encode(config),
	})
	require.NoError(t, err)
	require.Empty(t, formatDiagnostics(noImportResp.Diagnostics), "%s", formatDiagnostics(noImportResp.Diagnostics))
	require.Len(t, noImportResp.RequiresReplace, 1)
	assert.Equal(t, `AttributeName("cloud_subnet_cidr")`, noImportResp.RequiresReplace[0].String())
}

func formatDiagnostics(diags []*tfprotov6.Diagnostic) string {
	texts := make([]string, 0, len(diags))
	for _, d := range diags {
		texts = append(texts, d.Summary+": "+d.Detail)
	}

	return strings.Join(texts, "; ")
}

func TestClusterX509PendingKeyThroughProvider(t *testing.T) {
	fake := newFakeAPI(t)
	ctx := context.Background()
	server := configureClusterProvider(t, &testProviderConfig{client: fake.client(), project: testProject, pool: "ru-1"})
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	valueType := schemas.ResourceSchemas["selectel_mks_cluster_v2"].ValueType()
	encode := func(m clusterModel) *tfprotov6.DynamicValue {
		value, err := tfprotov6.NewDynamicValue(valueType, clusterStateOf(t, fake.resource(), m).Raw)
		require.NoError(t, err)

		return &value
	}
	privateValue := func(raw []byte, key string) (string, bool) {
		if len(raw) == 0 {
			return "", false
		}
		var keys map[string][]byte
		require.NoError(t, json.Unmarshal(raw, &keys))
		value, ok := keys[key]

		return string(value), ok
	}

	fake.fail(routePatchCluster, http.StatusConflict)
	plan := clusterPlan()
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringValue("Y2VydA=="),
	})
	applyResp, err := server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{
		TypeName: "selectel_mks_cluster_v2", Config: encode(plan), PlannedState: encode(plan),
	})
	require.NoError(t, err)
	require.Len(t, applyResp.Diagnostics, 1)
	assert.Equal(t, "X509 CA certificates of the cluster are not set", applyResp.Diagnostics[0].Summary)

	pending, ok := privateValue(applyResp.Private, x509PendingKey)
	require.True(t, ok, "the create must mark the pending x509: %s", applyResp.Private)
	assert.Equal(t, "true", pending)

	readResp, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{
		TypeName: "selectel_mks_cluster_v2", CurrentState: applyResp.NewState, Private: applyResp.Private,
	})
	require.NoError(t, err)
	require.Empty(t, formatDiagnostics(readResp.Diagnostics), "%s", formatDiagnostics(readResp.Diagnostics))

	read, err := readResp.NewState.Unmarshal(valueType)
	require.NoError(t, err)
	var readAttrs map[string]tftypes.Value
	require.NoError(t, read.As(&readAttrs))
	optionsValue := readAttrs["kubernetes_options"]
	var optionsAttrs map[string]tftypes.Value
	require.NoError(t, optionsValue.As(&optionsAttrs))
	require.True(t, optionsAttrs["x509_ca_certificates"].IsNull())

	fake.fail(routePatchCluster, 0)
	config := clusterState()
	config.ID = types.StringValue(testClusterID)
	config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringValue("Y2VydA=="),
	})
	updateResp, err := server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{
		TypeName: "selectel_mks_cluster_v2", Config: encode(config), PlannedState: encode(config),
		PriorState: readResp.NewState, PlannedPrivate: readResp.Private,
	})
	require.NoError(t, err)
	assert.Empty(t, formatDiagnostics(updateResp.Diagnostics), "%s", formatDiagnostics(updateResp.Diagnostics))

	_, ok = privateValue(updateResp.Private, x509PendingKey)
	assert.False(t, ok, "the update must clear the pending x509: %s", updateResp.Private)
	assert.Equal(t, "Y2VydA==", fake.storedX509(testClusterID))
}
