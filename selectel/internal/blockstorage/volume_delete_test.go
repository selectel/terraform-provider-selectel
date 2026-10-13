package blockstorage

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
)

func TestDeleteSafety(t *testing.T) {
	for _, scenario := range []string{"delete", "already deleting", "absent", "hidden", "attached", "in-use", "forbidden", "delete fails", "wait fails"} {
		t.Run(scenario, func(t *testing.T) {
			current := apiVolume()
			deleted := false
			deletes := 0
			gets := 0
			switch scenario {
			case "attached":
				current.Attachments = []volume.Attachment{{ID: volumeID}}
			case "in-use":
				current.Status = "in-use"
			case "already deleting":
				current.Status = "deleting"
			}
			client := testClient(t, volumeDeleteHandler(t, scenario, current, &deleted, &deletes, &gets))
			r := testResource(client)
			prior := model()
			if scenario == "delete" {
				prior.Timeouts = defaultTimeouts()
			}
			state := stateOf(t, r, prior)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			response := resource.DeleteResponse{State: state}
			r.Delete(ctx, resource.DeleteRequest{State: state}, &response)
			success := scenario == "delete" || scenario == "already deleting" || scenario == "absent"
			assert.Equal(t, !success, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			expected := 0
			if scenario == "delete" || scenario == "delete fails" || scenario == "wait fails" {
				expected = 1
			}
			assert.Equal(t, expected, deletes)
			if !success {
				expected := map[string]string{"hidden": "Confirm volume absence", "attached": "Volume is attached", "in-use": "Volume is attached", "forbidden": "Read volume before deletion", "delete fails": "Delete volume", "wait fails": "Wait for deleted volume"}
				assert.Equal(t, expected[scenario], response.Diagnostics.Errors()[0].Summary())
				assert.Equal(t, volumeID, readModel(t, response.State).ID.ValueString())
			}
		})
	}
}

func volumeDeleteHandler(
	t *testing.T,
	scenario string,
	current volume.View,
	deleted *bool,
	deletes, gets *int,
) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/volumes/detail" {
			all := []volume.View{}
			if scenario == "hidden" {
				all = append(all, current)
			}
			writeJSON(t, w, 200, map[string]any{"volumes": all})

			return
		}
		if req.Method == "DELETE" {
			(*deletes)++
			if scenario == "delete fails" {
				w.WriteHeader(409)
				return
			}
			*deleted = true
			w.WriteHeader(202)

			return
		}
		(*gets)++
		if scenario == "forbidden" {
			w.WriteHeader(403)
			return
		}
		if scenario == "wait fails" && *deleted {
			w.WriteHeader(500)
			return
		}
		if *deleted || scenario == "absent" || scenario == "hidden" || (scenario == "already deleting" && *gets > 1) {
			w.WriteHeader(404)
			return
		}
		writeJSON(t, w, 200, map[string]any{"volume": current})
	}
}
