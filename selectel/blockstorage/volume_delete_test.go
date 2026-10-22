package blockstorage

import (
	"cmp"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteSafety(t *testing.T) {
	attached := func(v *volume.View) { v.Attachments = []volume.Attachment{{ID: volumeID}} }
	for name, tc := range map[string]struct {
		volume func(*volume.View)
		// reads are the statuses of the volume reads in order, where 200
		// returns the volume; the last status repeats.
		reads []int
		// listed keeps the volume in the complete list after a 404.
		listed bool
		// deleteStatus answers the delete request; zero is 202.
		deleteStatus int
		deletes      int
		err          string
	}{
		"delete":                  {reads: []int{200, 404}, deletes: 1},
		"transient error in wait": {reads: []int{200, 503, 404}, deletes: 1},
		"already deleting":        {volume: func(v *volume.View) { v.Status = "deleting" }, reads: []int{200, 404}},
		"absent":                  {reads: []int{404}},
		"hidden":                  {reads: []int{404}, listed: true, err: "Confirm volume absence"},
		"attached":                {volume: attached, reads: []int{200}, err: "Volume is attached"},
		"in-use":                  {volume: func(v *volume.View) { v.Status = "in-use" }, reads: []int{200}, err: "Volume is attached"},
		"forbidden":               {reads: []int{403}, err: "Read volume before deletion"},
		"delete fails":            {reads: []int{200}, deleteStatus: 409, deletes: 1, err: "Delete volume"},
		"wait fails":              {reads: []int{200, 403}, deletes: 1, err: "Wait for deleted volume"},
	} {
		t.Run(name, func(t *testing.T) {
			current := apiVolume()
			if tc.volume != nil {
				tc.volume(&current)
			}
			reads, deletes := 0, 0
			r := testResource(testClient(t, func(w http.ResponseWriter, req *http.Request) {
				switch {
				case req.URL.Path == "/volumes/detail":
					all := []volume.View{}
					if tc.listed {
						all = append(all, current)
					}
					writeJSON(t, w, http.StatusOK, map[string]any{"volumes": all})
				case req.Method == http.MethodDelete:
					deletes++
					w.WriteHeader(cmp.Or(tc.deleteStatus, http.StatusAccepted))
				default:
					status := tc.reads[min(reads, len(tc.reads)-1)]
					reads++
					if status != http.StatusOK {
						w.WriteHeader(status)
						return
					}
					writeJSON(t, w, status, map[string]any{"volume": current})
				}
			}))
			prior := model()
			// The default timeouts apply when the configuration has no timeouts block.
			prior.Timeouts = defaultTimeouts()
			state := stateOf(t, r, prior)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			response := resource.DeleteResponse{State: state}
			r.Delete(ctx, resource.DeleteRequest{State: state}, &response)

			assert.Equal(t, tc.deletes, deletes)
			if tc.err == "" {
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
				return
			}
			require.True(t, response.Diagnostics.HasError())
			assert.Equal(t, tc.err, response.Diagnostics.Errors()[0].Summary())
			assert.Equal(t, volumeID, readModel(t, response.State).ID.ValueString())
		})
	}
}
