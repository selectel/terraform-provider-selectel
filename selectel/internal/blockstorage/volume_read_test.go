package blockstorage

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDriftAndAbsence(t *testing.T) {
	for _, scenario := range []string{"drift", "absent", "hidden", "forbidden", "list fails", "pending", "pending missing", "type uuid", "type alias", "type uuid changed", "type alias changed"} {
		t.Run(scenario, func(t *testing.T) {
			current := apiVolume()
			current.Size = 20
			current.Name = "external"
			current.Metadata = map[string]string{"team": "changed", createTokenKey: "token", "server": "keep"}
			current.Attachments = []volume.Attachment{{ID: volumeID, ServerID: typeID, Device: "/dev/vdb"}}
			client := testClient(t, volumeReadHandler(t, scenario, current))
			r := testResource(client)
			prior := model()
			prior.Metadata = metadata(map[string]string{"team": "old", "removed": "old"})
			if strings.HasPrefix(scenario, "pending") {
				prior.ID = types.StringValue("pending:token")
			}
			switch scenario {
			case "type uuid", "type uuid changed":
				prior.VolumeType = types.StringValue(typeID)
			case "type alias":
				prior.VolumeType = types.StringValue("fast.ru-1")
			case "type alias changed":
				prior.VolumeType = types.StringValue("basic.ru-1")
			}
			response := resource.ReadResponse{State: stateOf(t, r, prior)}
			r.Read(context.Background(), resource.ReadRequest{State: response.State}, &response)
			switch scenario {
			case "hidden", "forbidden", "list fails", "pending missing":
				require.True(t, response.Diagnostics.HasError())
				assert.Equal(t, prior.ID, readModel(t, response.State).ID)
				if scenario == "pending missing" {
					assert.Equal(t, "Recover volume", response.Diagnostics.Errors()[0].Summary())
					assert.Contains(t, response.Diagnostics.Errors()[0].Detail(), "not visible yet")
				}
			case "absent":
				require.False(t, response.Diagnostics.HasError())
				assert.True(t, response.State.Raw.IsNull())
			default:
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
				state := readModel(t, response.State)
				assert.Equal(t, int64(20), state.Size.ValueInt64())
				assert.Equal(t, "external", state.Name.ValueString())
				assert.Equal(t, metadata(map[string]string{"team": "changed"}), state.Metadata)
				assert.NotContains(t, state.ServiceMetadata.Elements(), createTokenKey)
				assert.Equal(t, volumeID, state.ID.ValueString())
				assert.Len(t, state.Attachments.Elements(), 1)
				switch scenario {
				case "type uuid", "type alias":
					assert.Equal(t, prior.VolumeType, state.VolumeType)
				case "type uuid changed", "type alias changed":
					assert.Equal(t, current.VolumeType, state.VolumeType.ValueString())
				}
			}
		})
	}
}

func volumeReadHandler(t *testing.T, scenario string, current volume.View) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/volumes/" + volumeID:
			switch scenario {
			case "absent", "hidden", "list fails":
				w.WriteHeader(404)
				return
			case "forbidden":
				w.WriteHeader(403)
				return
			}
			writeJSON(t, w, 200, map[string]any{"volume": current})
		case "/volumes/detail":
			all := []volume.View{}
			switch scenario {
			case "hidden", "pending":
				all = append(all, current)
			case "list fails":
				w.WriteHeader(403)
				return
			}
			writeJSON(t, w, 200, map[string]any{"volumes": all})
		case "/types/" + typeID:
			name := current.VolumeType
			if scenario == "type uuid changed" {
				name = "basic.ru-1a"
			}
			writeJSON(t, w, 200, map[string]any{"volume_type": map[string]any{"id": typeID, "name": name}})
		default:
			t.Errorf("unexpected request %s", req.URL.Path)
			w.WriteHeader(400)
		}
	}
}
