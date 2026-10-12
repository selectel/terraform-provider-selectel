package blockstorage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	volumeID = "11111111-1111-4111-8111-111111111111"
	typeID   = "22222222-2222-4222-8222-222222222222"
)

type testConfig struct {
	client          *sdk.Client
	err             error
	project, region string
}

func (c testConfig) BlockStorageClient(_ context.Context, _, _ string) (*sdk.Client, error) {
	return c.client, c.err
}
func (c testConfig) BlockStorageDefaults() (string, string) { return c.project, c.region }
func testClient(t *testing.T, handler http.HandlerFunc) *sdk.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "token", req.Header.Get("X-Auth-Token"))
		w.Header().Set("Content-Type", "application/json")
		handler(w, req)
	}))
	t.Cleanup(server.Close)
	client, err := sdk.NewClient(sdk.Config{Endpoint: server.URL, Token: "token"})
	require.NoError(t, err)

	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.WriteHeader(status)
	assert.NoError(t, json.NewEncoder(w).Encode(value))
}

func testResource(client *sdk.Client) *volumeResource {
	return &volumeResource{
		connection:   connection{config: testConfig{client: client, project: "project", region: "ru-1"}},
		pollInterval: time.Millisecond,
	}
}

func model() volumeModel {
	return volumeModel{
		ID:               types.StringValue(volumeID),
		ProjectID:        types.StringValue("project"),
		Region:           types.StringValue("ru-1"),
		Size:             types.Int64Value(10),
		Name:             types.StringValue("disk"),
		Description:      types.StringValue("description"),
		AvailabilityZone: types.StringValue("ru-1a"),
		VolumeType:       types.StringValue("fast.ru-1a"),
		SnapshotID:       types.StringValue(""),
		SourceVolumeID:   types.StringValue(""),
		Metadata:         types.MapNull(types.StringType),
		ServiceMetadata:  types.MapNull(types.StringType),
		Attachments:      types.SetNull(types.ObjectType{AttrTypes: attachmentTypes}),
		OnlineResize:     types.BoolValue(false),
		Timeouts: timeouts.Value{Object: types.ObjectValueMust(
			map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType},
			map[string]attr.Value{"create": types.StringValue("500ms"), "update": types.StringValue("500ms"), "delete": types.StringValue("500ms")},
		)},
	}
}

func apiVolume() volume.View {
	return volume.View{
		ID:               volumeID,
		Name:             "disk",
		Description:      "description",
		Size:             10,
		AvailabilityZone: "ru-1a",
		VolumeType:       "fast.ru-1a",
		Status:           "available",
		Metadata:         map[string]string{"server": "value"},
	}
}

func stateOf(t *testing.T, r *volumeResource, m volumeModel) tfsdk.State {
	t.Helper()
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	state := tfsdk.State{Schema: s.Schema}
	require.False(t, state.Set(context.Background(), &m).HasError())

	return state
}

func planOf(t *testing.T, r *volumeResource, m volumeModel) tfsdk.Plan {
	s := stateOf(t, r, m)
	return tfsdk.Plan(s)
}

func readModel(t *testing.T, state tfsdk.State) volumeModel {
	t.Helper()
	var m volumeModel
	require.False(t, state.Get(context.Background(), &m).HasError())

	return m
}

func metadata(values map[string]string) types.Map {
	value, d := types.MapValueFrom(context.Background(), types.StringType, values)
	if d.HasError() {
		panic(d)
	}

	return value
}

func defaultTimeouts() timeouts.Value {
	return timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType, "update": types.StringType, "delete": types.StringType,
	})}
}

// attributeErrors runs the validators of every root String and Map attribute
// against config, as Terraform does before calling ValidateConfig.
func attributeErrors[A any](t *testing.T, attributes map[string]A, config tfsdk.Config) diag.Diagnostics {
	t.Helper()
	ctx := context.Background()
	var diags diag.Diagnostics
	for name, attribute := range attributes {
		at := path.Root(name)
		if attribute, ok := any(attribute).(interface{ StringValidators() []validator.String }); ok {
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
		if attribute, ok := any(attribute).(interface{ MapValidators() []validator.Map }); ok {
			var value types.Map
			require.False(t, config.GetAttribute(ctx, at, &value).HasError())
			for _, rule := range attribute.MapValidators() {
				var response validator.MapResponse
				rule.ValidateMap(ctx, validator.MapRequest{
					Path: at, PathExpression: at.Expression(), ConfigValue: value, Config: config,
				}, &response)
				diags.Append(response.Diagnostics...)
			}
		}
	}

	return diags
}

func volumeAttributeErrors(t *testing.T, m volumeModel) diag.Diagnostics {
	t.Helper()
	r := testResource(nil)
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)

	return attributeErrors(t, s.Schema.Attributes, tfsdk.Config(stateOf(t, r, m)))
}
