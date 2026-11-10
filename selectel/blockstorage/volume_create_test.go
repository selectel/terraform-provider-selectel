package blockstorage

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRequestAndState(t *testing.T) {
	for _, source := range []string{"empty", "snapshot", "volume", "image", "backup"} {
		t.Run(source, func(t *testing.T) {
			var posted map[string]any
			gets := 0
			current := apiVolume()
			client := testClient(t, func(w http.ResponseWriter, req *http.Request) {
				switch req.Method + " " + req.URL.Path {
				case "POST /volumes":
					var body map[string]map[string]any
					assert.NoError(t, json.NewDecoder(req.Body).Decode(&body))
					posted = body["volume"]
					if source == "backup" {
						assert.Contains(t, req.Header.Get("OpenStack-API-Version"), "3.47")
					}
					writeJSON(t, w, 202, map[string]any{"volume": map[string]any{"id": volumeID, "status": "creating"}})
				case "GET /volumes/" + volumeID:
					gets++
					v := current
					if gets == 1 {
						v.Status = "creating"
					}
					writeJSON(t, w, 200, map[string]any{"volume": v})
				default:
					t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
					w.WriteHeader(400)
				}
			})
			r := testResource(client)
			plan := model()
			plan.Metadata = metadata(map[string]string{"team": "infra"})
			current.Metadata["team"] = "infra"
			switch source {
			case "snapshot":
				plan.SnapshotID = types.StringValue(typeID)
				current.SnapshotID = typeID
				plan.Size = types.Int64Unknown()
			case "volume":
				plan.SourceVolumeID = types.StringValue(typeID)
				current.SourceVolID = typeID
				plan.Size = types.Int64Unknown()
			case "image":
				plan.ImageID = types.StringValue(typeID)
			case "backup":
				plan.BackupID = types.StringValue(typeID)
				plan.Size = types.Int64Unknown()
			}
			if source == "empty" {
				plan.Timeouts = defaultTimeouts()
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			response := resource.CreateResponse{State: stateOf(t, r, plan)}
			r.Create(ctx, resource.CreateRequest{Plan: planOf(t, r, plan)}, &response)
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			assert.Equal(t, "disk", posted["name"])
			assert.Equal(t, "description", posted["description"])
			assert.Equal(t, "ru-1a", posted["availability_zone"])
			assert.Equal(t, "fast.ru-1a", posted["volume_type"])
			assert.Equal(t, "infra", posted["metadata"].(map[string]any)["team"])
			assert.NotEmpty(t, posted["metadata"].(map[string]any)[createTokenKey])
			if plan.Size.IsUnknown() {
				assert.NotContains(t, posted, "size")
			} else {
				assert.Equal(t, float64(10), posted["size"])
			}
			key := map[string]string{"snapshot": "snapshot_id", "volume": "source_volid", "image": "imageRef", "backup": "backup_id"}[source]
			if key != "" {
				assert.Equal(t, typeID, posted[key])
			}
			state := readModel(t, response.State)
			assert.Equal(t, volumeID, state.ID.ValueString())
			assert.Equal(t, int64(10), state.Size.ValueInt64())
			assert.Equal(t, plan.Metadata, state.Metadata)
			assert.Equal(t, metadata(current.Metadata), state.ServiceMetadata)
			assert.Equal(t, 2, gets)
		})
	}
}

func TestCreateFailureKeepsRecoverableIdentity(t *testing.T) {
	for _, tc := range []struct {
		name       string
		postStatus int  // 0 accepts the request
		postNoID   bool // accepted, but the response lacks an ID
		postHang   bool // no response before the deadline
		getStatus  string
		timeout    bool
		want       string // "nothing", "pending" or "id" in state
	}{
		{name: "rejected", postStatus: 400, want: "nothing"},
		{name: "missing source", postStatus: 404, want: "nothing"},
		{name: "quota", postStatus: 413, want: "nothing"},
		{name: "rate limited", postStatus: 429, want: "nothing"},
		{name: "uncertain", postStatus: 500, want: "pending"},
		{name: "missing id", postNoID: true, want: "pending"},
		{name: "no response", postHang: true, timeout: true, want: "pending"},
		{name: "terminal", getStatus: "error", want: "id"},
		{name: "timeout", getStatus: "creating", timeout: true, want: "id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			posts := 0
			var token string
			current := apiVolume()
			client := testClient(t, func(w http.ResponseWriter, req *http.Request) {
				switch {
				case req.Method == "POST":
					posts++
					var body map[string]map[string]any
					assert.NoError(t, json.NewDecoder(req.Body).Decode(&body))
					token, _ = body["volume"]["metadata"].(map[string]any)[createTokenKey].(string)
					switch {
					case tc.postStatus != 0:
						writeJSON(t, w, tc.postStatus, map[string]any{})
					case tc.postHang:
						<-req.Context().Done()
					case tc.postNoID:
						writeJSON(t, w, 202, map[string]any{"volume": map[string]any{"status": "creating"}})
					default:
						writeJSON(t, w, 202, map[string]any{"volume": current})
					}
				case req.URL.Path == "/volumes/detail":
					// The volume that the uncertain POST created is found by its token.
					found := current
					found.Metadata = map[string]string{createTokenKey: token}
					writeJSON(t, w, 200, map[string]any{"volumes": []volume.View{found}})
				default:
					v := current
					if tc.getStatus != "" {
						v.Status = tc.getStatus
					}
					writeJSON(t, w, 200, map[string]any{"volume": v})
				}
			})
			r := testResource(client)
			plan := model()
			plan.ID = types.StringUnknown()
			plan.VolumeType = types.StringValue("fast.ru-1") // a regional alias of fast.ru-1a
			response := resource.CreateResponse{State: tfsdk.State{Schema: stateOf(t, r, plan).Schema}}
			ctx := context.Background()
			if tc.timeout {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 20*time.Millisecond)
				defer cancel()
			}
			r.Create(ctx, resource.CreateRequest{Plan: planOf(t, r, plan)}, &response)
			require.True(t, response.Diagnostics.HasError())
			assert.Equal(t, 1, posts)
			if tc.want == "nothing" {
				assert.True(t, response.State.Raw.IsNull())
				return
			}
			state := readModel(t, response.State)
			// The configured type is kept until Read verifies it against the API name.
			assert.Equal(t, "fast.ru-1", state.VolumeType.ValueString())
			if tc.want == "id" {
				assert.Equal(t, volumeID, state.ID.ValueString())
				return
			}
			require.NotEmpty(t, token)
			assert.Equal(t, "pending:"+token, state.ID.ValueString())

			read := resource.ReadResponse{State: response.State}
			r.Read(context.Background(), resource.ReadRequest{State: response.State}, &read)
			require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
			recovered := readModel(t, read.State)
			assert.Equal(t, volumeID, recovered.ID.ValueString())
			assert.Equal(t, "fast.ru-1", recovered.VolumeType.ValueString())
			assert.Equal(t, 1, posts)
		})
	}
}
