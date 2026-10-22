package blockstorage

import (
	"cmp"
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDriftAndAbsence(t *testing.T) {
	current := apiVolume()
	current.Size = 20
	current.Name = "external"
	current.Metadata = map[string]string{"team": "changed", createTokenKey: "token", "server": "keep"}
	current.Attachments = []volume.Attachment{{ID: volumeID, ServerID: typeID, Device: "/dev/vdb"}}
	// Empty fields keep the values of model and apiVolume; a zero status is 200.
	for name, tc := range map[string]struct {
		priorID   string
		priorType string
		// get and list are the statuses of the volume read and of the
		// complete list.
		get, list int
		listed    bool
		// typeName is the name that the API reports for the volume type ID.
		typeName string
		gone     bool
		err      string
		// stateType is the volume type in the refreshed state.
		stateType string
	}{
		"drift":              {stateType: "fast.ru-1a"},
		"absent":             {get: 404, gone: true},
		"hidden":             {get: 404, listed: true, err: "Confirm volume absence"},
		"forbidden":          {get: 403, err: "Read volume"},
		"list fails":         {get: 404, list: 403, err: "Confirm volume absence"},
		"pending":            {priorID: "pending:token", listed: true, stateType: "fast.ru-1a"},
		"pending missing":    {priorID: "pending:token", err: "Recover volume"},
		"type uuid":          {priorType: typeID, stateType: typeID},
		"type alias":         {priorType: "fast.ru-1", stateType: "fast.ru-1"},
		"type uuid changed":  {priorType: typeID, typeName: "basic.ru-1a", stateType: current.VolumeType},
		"type alias changed": {priorType: "basic.ru-1", stateType: current.VolumeType},
	} {
		t.Run(name, func(t *testing.T) {
			r := testResource(testClient(t, func(w http.ResponseWriter, req *http.Request) {
				switch req.URL.Path {
				case "/volumes/" + volumeID:
					if tc.get != 0 {
						w.WriteHeader(tc.get)
						return
					}
					writeJSON(t, w, http.StatusOK, map[string]any{"volume": current})
				case "/volumes/detail":
					if tc.list != 0 {
						w.WriteHeader(tc.list)
						return
					}
					all := []volume.View{}
					if tc.listed {
						all = append(all, current)
					}
					writeJSON(t, w, http.StatusOK, map[string]any{"volumes": all})
				case "/types/" + typeID:
					writeJSON(t, w, http.StatusOK, map[string]any{"volume_type": map[string]any{"id": typeID, "name": cmp.Or(tc.typeName, current.VolumeType)}})
				default:
					t.Errorf("unexpected request %s", req.URL.Path)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			prior := model()
			prior.ID = types.StringValue(cmp.Or(tc.priorID, volumeID))
			prior.VolumeType = types.StringValue(cmp.Or(tc.priorType, prior.VolumeType.ValueString()))
			prior.Metadata = metadata(map[string]string{"team": "old", "removed": "old"})
			response := resource.ReadResponse{State: stateOf(t, r, prior)}
			r.Read(context.Background(), resource.ReadRequest{State: response.State}, &response)

			if tc.err != "" {
				require.True(t, response.Diagnostics.HasError())
				assert.Equal(t, tc.err, response.Diagnostics.Errors()[0].Summary())
				assert.Equal(t, prior.ID, readModel(t, response.State).ID)

				return
			}
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			if tc.gone {
				assert.True(t, response.State.Raw.IsNull())
				return
			}
			state := readModel(t, response.State)
			assert.Equal(t, volumeID, state.ID.ValueString())
			assert.Equal(t, int64(20), state.Size.ValueInt64())
			assert.Equal(t, "external", state.Name.ValueString())
			assert.Equal(t, metadata(map[string]string{"team": "changed"}), state.Metadata)
			assert.NotContains(t, state.ServiceMetadata.Elements(), createTokenKey)
			assert.Len(t, state.Attachments.Elements(), 1)
			assert.Equal(t, tc.stateType, state.VolumeType.ValueString())
		})
	}
}
