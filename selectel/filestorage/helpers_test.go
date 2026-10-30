package filestorage

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/stretchr/testify/require"
)

const (
	shareID   = "11111111-1111-4111-8111-111111111111"
	networkID = "22222222-2222-4222-8222-222222222222"
	subnetID  = "33333333-3333-4333-8333-333333333333"
)

type testConfig struct {
	client          *sdk.Client
	err             error
	project, region string
}

func (c testConfig) FileStorageClient(_ context.Context, _, _ string) (*sdk.Client, error) {
	return c.client, c.err
}

func (c testConfig) FileStorageDefaults() (string, string) { return c.project, c.region }

func testShareResource(client *sdk.Client) *shareResource {
	return &shareResource{
		connection:     connection{config: testConfig{client: client, project: "project", region: "ru-9"}},
		pollInterval:   time.Millisecond,
		recoveryWindow: 50 * time.Millisecond,
	}
}

func shareState() shareModel {
	return shareModel{
		ID:               types.StringValue(shareID),
		ProjectID:        types.StringValue("project"),
		Region:           types.StringValue("ru-9"),
		ShareProto:       types.StringValue(protoNFS),
		Size:             types.Int64Value(50),
		ShareType:        types.StringValue("basic.ru-9a"),
		NetworkID:        types.StringValue(networkID),
		SubnetID:         types.StringValue(subnetID),
		IP:               types.StringValue("10.0.0.5"),
		Name:             types.StringValue("data"),
		Description:      types.StringValue(""),
		Metadata:         types.MapNull(types.StringType),
		AllMetadata:      metadata(map[string]string{"delete_empty_share_network": "enabled"}),
		Status:           types.StringValue("available"),
		AvailabilityZone: types.StringValue("ru-9a"),
		ExportLocations: types.ListValueMust(types.ObjectType{AttrTypes: exportLocationTypes}, []attr.Value{
			types.ObjectValueMust(exportLocationTypes, map[string]attr.Value{
				"path":      types.StringValue("10.0.0.5:/shares/share-1"),
				"preferred": types.BoolValue(true),
			}),
		}),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType, "update": types.StringType, "delete": types.StringType,
		})},
	}
}

func shareStateOf(t *testing.T, r *shareResource, m shareModel) tfsdk.State {
	t.Helper()
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	state := tfsdk.State{Schema: s.Schema}
	require.False(t, state.Set(context.Background(), &m).HasError())

	return state
}

func metadata(values map[string]string) types.Map {
	value, d := types.MapValueFrom(context.Background(), types.StringType, values)
	if d.HasError() {
		panic(d)
	}

	return value
}

// attributeErrors runs the validators of every root String and Int64 attribute
// against config, as Terraform does before calling ValidateConfig.
func attributeErrors(t *testing.T, attributes map[string]schema.Attribute, config tfsdk.Config) diag.Diagnostics {
	t.Helper()
	ctx := context.Background()
	var diags diag.Diagnostics
	for name, attribute := range attributes {
		at := path.Root(name)
		if attribute, ok := attribute.(interface{ StringValidators() []validator.String }); ok {
			var value types.String
			require.False(t, config.GetAttribute(ctx, at, &value).HasError())
			for _, rule := range attribute.StringValidators() {
				var response validator.StringResponse
				rule.ValidateString(ctx, validator.StringRequest{
					Path: at, PathExpression: at.Expression(), ConfigValue: value, Config: config,
				}, &response)
				diags.Append(response.Diagnostics...)
			}
		}
		if attribute, ok := attribute.(interface{ Int64Validators() []validator.Int64 }); ok {
			var value types.Int64
			require.False(t, config.GetAttribute(ctx, at, &value).HasError())
			for _, rule := range attribute.Int64Validators() {
				var response validator.Int64Response
				rule.ValidateInt64(ctx, validator.Int64Request{
					Path: at, PathExpression: at.Expression(), ConfigValue: value, Config: config,
				}, &response)
				diags.Append(response.Diagnostics...)
			}
		}
	}

	return diags
}

type validatorResponse struct{ validator.StringResponse }

func stringRequest(value string) validator.StringRequest {
	at := path.Root("value")

	return validator.StringRequest{Path: at, PathExpression: at.Expression(), ConfigValue: types.StringValue(value)}
}
