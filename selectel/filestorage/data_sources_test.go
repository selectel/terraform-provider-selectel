package filestorage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readData runs a data source with a configuration that sets the given
// attributes and leaves the others null.
func readData[M any](t *testing.T, d datasource.DataSource, client *sdk.Client, config M) (M, diag.Diagnostics) {
	t.Helper()
	ctx := context.Background()
	var configured datasource.ConfigureResponse
	d.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{
		ProviderData: testConfig{client: client, project: "project", region: "ru-9"},
	}, &configured)
	require.False(t, configured.Diagnostics.HasError())
	var s datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &s)
	state := tfsdk.State{Schema: s.Schema}
	require.False(t, state.Set(ctx, &config).HasError())
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: s.Schema, Raw: tftypes.NewValue(state.Raw.Type(), nil)}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config(state)}, &resp)
	var got M
	if !resp.State.Raw.IsNull() {
		require.False(t, resp.State.Get(ctx, &got).HasError())
	}

	return got, resp.Diagnostics
}

func shareSelector() shareDataModel {
	return shareDataModel{
		ProjectID:          types.StringValue("project"),
		Region:             types.StringValue("ru-9"),
		ShareID:            types.StringNull(),
		Name:               types.StringNull(),
		Description:        types.StringNull(),
		Status:             types.StringNull(),
		Metadata:           types.MapNull(types.StringType),
		ExportLocationPath: types.StringNull(),
		ShareProto:         types.StringNull(),
		ShareType:          types.StringNull(),
		Size:               types.Int64Null(),
		NetworkID:          types.StringNull(),
		SubnetID:           types.StringNull(),
		IP:                 types.StringNull(),
		AllMetadata:        types.MapNull(types.StringType),
		AvailabilityZone:   types.StringNull(),
		ExportLocations:    types.ListNull(types.ObjectType{AttrTypes: exportLocationTypes}),
	}
}

func otherShare(id, name string) map[string]any {
	return map[string]any{"id": id, "name": name, "status": "available", "metadata": map[string]any{}}
}

func TestShareDataSource(t *testing.T) {
	t.Run("one match by name", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share["metadata"].(map[string]any)["owner"] = "team"
		fake.others = []map[string]any{otherShare(networkID, "other")}
		config := shareSelector()
		config.Name = types.StringValue("data")

		got, diags := readData(t, NewShareDataSource(), fake.client(), config)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, shareID, got.ID.ValueString())
		assert.Equal(t, networkID, got.NetworkID.ValueString())
		assert.Equal(t, subnetID, got.SubnetID.ValueString())
		assert.Equal(t, "10.0.0.5", got.IP.ValueString())
		assert.Equal(t, "basic.ru-9a", got.ShareType.ValueString())
		assert.Equal(t, metadata(map[string]string{"owner": "team"}), got.Metadata)
		assert.Contains(t, got.AllMetadata.Elements(), deleteNetworkKey)
		assert.Len(t, got.ExportLocations.Elements(), 1)
	})

	t.Run("by share_id without a list request", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share["metadata"].(map[string]any)["owner"] = "team"
		config := shareSelector()
		config.ShareID = types.StringValue(shareID)

		got, diags := readData(t, NewShareDataSource(), fake.client(), config)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, shareID, got.ID.ValueString())
		assert.Equal(t, shareID, got.ShareID.ValueString())
		assert.Equal(t, "data", got.Name.ValueString())
		assert.Equal(t, "10.0.0.5", got.IP.ValueString())
		assert.Equal(t, metadata(map[string]string{"owner": "team"}), got.Metadata)
		assert.Zero(t, fake.count("GET /shares/detail"))
	})

	t.Run("a missing share_id is not found", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.fail["GET /shares/"+shareID] = answer{
			http.StatusNotFound, `{"itemNotFound": {"code": 404, "message": "Share could not be found."}}`,
		}
		config := shareSelector()
		config.ShareID = types.StringValue(shareID)

		_, diags := readData(t, NewShareDataSource(), fake.client(), config)

		require.True(t, diags.HasError())
		assert.Equal(t, "File storage not found", diags.Errors()[0].Summary())
		assert.Contains(t, diags.Errors()[0].Detail(), shareID)
	})

	t.Run("one match by mount path", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.others = []map[string]any{otherShare(networkID, "data")}
		config := shareSelector()
		config.ExportLocationPath = types.StringValue("10.0.0.5:/shares/share-1")

		got, diags := readData(t, NewShareDataSource(), fake.client(), config)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, shareID, got.ID.ValueString())
		assert.Equal(t, "10.0.0.5:/shares/share-1", fake.requestQuery("GET /shares/detail", "export_location_path"))
	})

	t.Run("one match by a service key", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.others = []map[string]any{otherShare(networkID, "data")}
		config := shareSelector()
		config.Metadata = metadata(map[string]string{deleteNetworkKey: "enabled"})

		got, diags := readData(t, NewShareDataSource(), fake.client(), config)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, shareID, got.ID.ValueString())
		assert.Equal(t, config.Metadata, got.Metadata, "a configured selector stays as configured")
	})

	t.Run("no match", func(t *testing.T) {
		fake := newFakeAPI(t)
		config := shareSelector()
		config.Name = types.StringValue("missing")

		_, diags := readData(t, NewShareDataSource(), fake.client(), config)

		require.True(t, diags.HasError())
		assert.Equal(t, "File storage not found", diags.Errors()[0].Summary())
	})

	t.Run("several matches", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.others = []map[string]any{otherShare(networkID, "data")}
		config := shareSelector()
		config.Name = types.StringValue("data")

		_, diags := readData(t, NewShareDataSource(), fake.client(), config)

		require.True(t, diags.HasError())
		assert.Equal(t, "Multiple file storages found", diags.Errors()[0].Summary())
	})

	for name, failure := range map[string]answer{
		"forbidden":    {http.StatusForbidden, `{"forbidden": {"code": 403, "message": "Policy does not allow."}}`},
		"server error": {http.StatusInternalServerError, ""},
	} {
		t.Run(name+" is not a missing share", func(t *testing.T) {
			fake := newFakeAPI(t)
			fake.fail["GET /shares/detail"] = failure

			_, diags := readData(t, NewShareDataSource(), fake.client(), shareSelector())

			require.True(t, diags.HasError())
			assert.Equal(t, "Find file storage", diags.Errors()[0].Summary())
		})
	}
}

func serveJSON(t *testing.T, path, body string) *sdk.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != path {
			t.Errorf("unexpected request %s", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	client, err := sdk.NewClient(sdk.Config{Endpoint: server.URL, Token: "token"})
	require.NoError(t, err)

	return client
}

func TestAvailabilityZonesDataSource(t *testing.T) {
	client := serveJSON(t, "/availability-zones", `{"availability_zones": [
		{"id": "az-2", "name": "ru-9b"}, {"id": "az-1", "name": "ru-9a"}
	]}`)
	config := availabilityZonesDataModel{
		ProjectID: types.StringValue("project"), Region: types.StringValue("ru-9"),
		ID: types.StringNull(), Names: types.ListNull(types.StringType),
	}

	got, diags := readData(t, NewAvailabilityZonesDataSource(), client, config)

	require.False(t, diags.HasError(), "%v", diags)
	var names []string
	require.False(t, got.Names.ElementsAs(context.Background(), &names, false).HasError())
	assert.Equal(t, []string{"ru-9a", "ru-9b"}, names)
	assert.Equal(t, "ru-9", got.ID.ValueString())
}
