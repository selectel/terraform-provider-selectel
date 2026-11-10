package images

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/images-go/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	imageID   = "11111111-1111-4111-8111-111111111111"
	projectID = "a07abc12310546f1b9291ab3013a7d75"
	memberID  = "b07abc12310546f1b9291ab3013a7d75"
)

type testConfig struct{ client *sdk.Client }

func (c testConfig) ImagesClient(context.Context, string, string) (*sdk.Client, error) {
	return c.client, nil
}
func (c testConfig) ImagesDefaults() (string, string) { return projectID, "ru-6" }

// testConnection serves the API with handler through the real images-go client.
func testConnection(t *testing.T, handler http.HandlerFunc) connection {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "token", req.Header.Get("X-Auth-Token"))
		w.Header().Set("Content-Type", "application/json")
		handler(w, req)
	}))
	t.Cleanup(server.Close)
	client, err := sdk.NewClient(sdk.Config{Endpoint: server.URL, Token: "token"})
	require.NoError(t, err)

	return connection{config: testConfig{client: client}}
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.WriteHeader(status)
	assert.NoError(t, json.NewEncoder(w).Encode(value))
}

// apiImage is an image as the API reports it, with fields replaced by changes.
func apiImage(changes map[string]any) map[string]any {
	view := map[string]any{
		"id": imageID, "name": "image", "status": "active", "visibility": "private",
		"protected": false, "os_hidden": false, "owner": projectID, "size": 3,
		"container_format": "bare", "disk_format": "raw", "min_disk": 1, "min_ram": 0,
		"created_at": "2026-10-09T10:00:00Z", "updated_at": "2026-10-09T10:00:00Z",
		"tags": []string{}, "os_distro": "ubuntu",
	}
	maps.Copy(view, changes)

	return view
}

func testImageResource(t *testing.T, handler http.HandlerFunc) *imageResource {
	t.Helper()
	return &imageResource{connection: testConnection(t, handler), pollInterval: time.Millisecond}
}

// planned is the image configuration that the tests start from.
func planned() imageModel {
	return imageModel{
		ID:                types.StringUnknown(),
		ProjectID:         types.StringValue(projectID),
		Region:            types.StringValue("ru-6"),
		Name:              types.StringValue("image"),
		ContainerFormat:   types.StringValue("bare"),
		DiskFormat:        types.StringValue("raw"),
		LocalFilePath:     types.StringNull(),
		ImageSourceURL:    types.StringNull(),
		MinDiskGB:         types.Int64Unknown(),
		MinRAMMB:          types.Int64Value(0),
		Protected:         types.BoolValue(false),
		Hidden:            types.BoolValue(false),
		Visibility:        types.StringValue("private"),
		Tags:              types.SetNull(types.StringType),
		Properties:        types.MapNull(types.StringType),
		ServiceProperties: types.MapUnknown(types.StringType),
		Status:            types.StringUnknown(),
		Owner:             types.StringUnknown(),
		SizeBytes:         types.Int64Unknown(),
		Checksum:          types.StringUnknown(),
		CreatedAt:         types.StringUnknown(),
		UpdatedAt:         types.StringUnknown(),
		Timeouts: timeouts.Value{Object: types.ObjectValueMust(
			map[string]attr.Type{"create": types.StringType, "delete": types.StringType},
			map[string]attr.Value{"create": types.StringValue("1s"), "delete": types.StringValue("1s")},
		)},
	}
}

func properties(values map[string]string) types.Map {
	return types.MapValueMust(types.StringType, stringValues(values))
}

func stringValues(values map[string]string) map[string]attr.Value {
	result := map[string]attr.Value{}
	for key, value := range values {
		result[key] = types.StringValue(value)
	}

	return result
}

func imageState(t *testing.T, r *imageResource, m imageModel) tfsdk.State {
	t.Helper()
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	state := tfsdk.State{Schema: s.Schema}
	require.False(t, state.Set(context.Background(), &m).HasError())

	return state
}

func readImage(t *testing.T, state tfsdk.State) imageModel {
	t.Helper()
	var m imageModel
	require.False(t, state.Get(context.Background(), &m).HasError())

	return m
}
