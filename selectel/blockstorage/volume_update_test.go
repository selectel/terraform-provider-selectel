package blockstorage

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFields(t *testing.T) {
	for _, tc := range []struct {
		name    string
		change  func(*volumeModel)
		request map[string]any
	}{
		{"name", func(m *volumeModel) { m.Name = types.StringValue("new") }, map[string]any{"name": "new"}},
		{"description", func(m *volumeModel) { m.Description = types.StringValue("new description") }, map[string]any{"description": "new description"}},
		{"clear fields", func(m *volumeModel) { m.Name = types.StringValue(""); m.Description = types.StringValue("") }, map[string]any{"name": "", "description": ""}},
		{
			"metadata", func(m *volumeModel) { m.Metadata = metadata(map[string]string{"team": "new", iopsKey: "3000"}) },
			map[string]any{"metadata": map[string]any{"team": "new", iopsKey: "3000", bandwidthKey: "123", createTokenKey: "token"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current := apiVolume()
			current.Metadata = map[string]string{"team": "old", "remove": "old", iopsKey: "2000", bandwidthKey: "123", createTokenKey: "token"}
			var request map[string]any
			client := testClient(t, func(w http.ResponseWriter, req *http.Request) {
				switch req.Method {
				case "GET":
					writeJSON(t, w, 200, map[string]any{"volume": current})
				case "PUT":
					var body struct {
						Volume json.RawMessage `json:"volume"`
					}
					assert.NoError(t, json.NewDecoder(req.Body).Decode(&body))
					assert.NoError(t, json.Unmarshal(body.Volume, &request))
					if _, ok := request["metadata"]; ok {
						current.Metadata = nil
					}
					assert.NoError(t, json.Unmarshal(body.Volume, &current))
					writeJSON(t, w, 200, map[string]any{"volume": current})
				default:
					t.Errorf("unexpected %s request", req.Method)
					w.WriteHeader(400)
				}
			})
			r := testResource(client)
			prior := model()
			prior.Metadata = metadata(map[string]string{"team": "old", "remove": "old", iopsKey: "2000"})
			plan := prior
			tc.change(&plan)
			response := resource.UpdateResponse{State: stateOf(t, r, prior)}
			r.Update(t.Context(), resource.UpdateRequest{State: response.State, Plan: planOf(t, r, plan)}, &response)
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			assert.Equal(t, tc.request, request)
			state := readModel(t, response.State)
			assert.Equal(t, plan.Name, state.Name)
			assert.Equal(t, plan.Description, state.Description)
			assert.Equal(t, plan.Metadata, state.Metadata)
			assert.Equal(t, prior.ID, state.ID)
		})
	}
}

func TestUpdateResize(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		remoteSize, target     int
		attached, online       bool
		readError, extendError bool
		finalStatus            string
		wantError              bool
		wantExtends, wantSize  int
	}{
		{name: "grow", remoteSize: 10, target: 20, wantExtends: 1, wantSize: 20},
		{name: "online", remoteSize: 10, target: 20, attached: true, online: true, wantExtends: 1, wantSize: 20},
		{name: "attached blocked", remoteSize: 10, target: 20, attached: true, wantError: true, wantSize: 10},
		{name: "shrink", remoteSize: 10, target: 5, wantError: true, wantSize: 10},
		{name: "read fails", remoteSize: 10, target: 20, readError: true, wantError: true, wantSize: 10},
		{name: "resize fails", remoteSize: 10, target: 20, extendError: true, wantError: true, wantExtends: 1, wantSize: 10},
		{name: "wait fails", remoteSize: 10, target: 20, finalStatus: "error_extending", wantError: true, wantExtends: 1, wantSize: 10},
		{name: "already resized", remoteSize: 20, target: 20, wantSize: 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current := apiVolume()
			current.Size = tc.remoteSize
			if tc.attached {
				current.Status = "in-use"
			}
			extends := 0
			client := testClient(t, func(w http.ResponseWriter, req *http.Request) {
				switch req.Method {
				case "GET":
					if tc.readError {
						w.WriteHeader(503)
						return
					}
					writeJSON(t, w, 200, map[string]any{"volume": current})
				case "POST":
					extends++
					var body map[string]map[string]int
					assert.NoError(t, json.NewDecoder(req.Body).Decode(&body))
					assert.Equal(t, tc.target, body["os-extend"]["new_size"])
					if tc.online {
						assert.Contains(t, req.Header.Get("OpenStack-API-Version"), "3.42")
					}
					if tc.extendError {
						w.WriteHeader(500)
						return
					}
					if tc.finalStatus != "" {
						current.Status = tc.finalStatus
					} else {
						current.Size = tc.target
					}
					w.WriteHeader(202)
				default:
					t.Errorf("unexpected %s request", req.Method)
					w.WriteHeader(400)
				}
			})
			r := testResource(client)
			prior := model()
			plan := prior
			plan.Size = types.Int64Value(int64(tc.target))
			plan.OnlineResize = types.BoolValue(tc.online)
			response := resource.UpdateResponse{State: stateOf(t, r, prior)}
			r.Update(t.Context(), resource.UpdateRequest{State: response.State, Plan: planOf(t, r, plan)}, &response)
			assert.Equal(t, tc.wantError, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			assert.Equal(t, tc.wantExtends, extends)
			state := readModel(t, response.State)
			assert.Equal(t, int64(tc.wantSize), state.Size.ValueInt64())
			assert.Equal(t, prior.ID, state.ID)
		})
	}
}

func TestUpdatePartialSuccessAndRetry(t *testing.T) {
	for _, tc := range []struct {
		name         string
		rejectUpdate bool
	}{
		{"update rejected", true}, {"resize rejected after update", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current := apiVolume()
			// total_iops_sec has a server default, which the plan changes.
			current.Metadata = map[string]string{"team": "old", "remove": "old", iopsKey: "500", createTokenKey: "token"}
			failNext := true
			client := testClient(t, func(w http.ResponseWriter, req *http.Request) {
				switch req.Method {
				case "GET":
					writeJSON(t, w, 200, map[string]any{"volume": current})
				case "PUT":
					if tc.rejectUpdate && failNext {
						failNext = false
						w.WriteHeader(500)
						return
					}
					var body struct {
						Volume json.RawMessage `json:"volume"`
					}
					assert.NoError(t, json.NewDecoder(req.Body).Decode(&body))
					var fields map[string]json.RawMessage
					assert.NoError(t, json.Unmarshal(body.Volume, &fields))
					if _, ok := fields["metadata"]; ok {
						current.Metadata = nil
					}
					assert.NoError(t, json.Unmarshal(body.Volume, &current))
					writeJSON(t, w, 200, map[string]any{"volume": current})
				case "POST":
					if failNext {
						failNext = false
						w.WriteHeader(500)
						return
					}
					current.Size = 20
					w.WriteHeader(202)
				default:
					t.Errorf("unexpected %s request", req.Method)
					w.WriteHeader(400)
				}
			})
			r := testResource(client)
			prior := model()
			prior.Metadata = metadata(map[string]string{"team": "old", "remove": "old"})
			plan := prior
			plan.Name = types.StringValue("new")
			plan.Metadata = metadata(map[string]string{"team": "new", iopsKey: "1000"})
			plan.Size = types.Int64Value(20)
			response := resource.UpdateResponse{State: stateOf(t, r, prior)}
			r.Update(t.Context(), resource.UpdateRequest{State: response.State, Plan: planOf(t, r, plan)}, &response)
			require.True(t, response.Diagnostics.HasError())
			state := readModel(t, response.State)
			assert.Equal(t, prior.ID, state.ID)
			assert.Equal(t, prior.Size, state.Size)
			if tc.rejectUpdate {
				assert.Equal(t, prior.Name, state.Name)
				assert.Equal(t, prior.Metadata, state.Metadata)
			} else {
				assert.Equal(t, plan.Name, state.Name)
				assert.Equal(t, plan.Metadata, state.Metadata)
			}
			retry := resource.UpdateResponse{State: response.State}
			r.Update(t.Context(), resource.UpdateRequest{State: response.State, Plan: planOf(t, r, plan)}, &retry)
			require.False(t, retry.Diagnostics.HasError(), "%v", retry.Diagnostics)
			final := readModel(t, retry.State)
			assert.Equal(t, plan.Name, final.Name)
			assert.Equal(t, plan.Metadata, final.Metadata)
			assert.Equal(t, plan.Size, final.Size)
			assert.NotContains(t, current.Metadata, "remove")
			assert.Equal(t, "token", current.Metadata[createTokenKey])
		})
	}
}
