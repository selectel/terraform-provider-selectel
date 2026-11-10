package filestorage

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func shortTimeouts(duration string) timeouts.Value {
	return timeouts.Value{Object: types.ObjectValueMust(
		map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType},
		map[string]attr.Value{
			"create": types.StringValue(duration), "update": types.StringValue(duration), "delete": types.StringValue(duration),
		},
	)}
}

// createPlan is the plan of a new share: the computed values are unknown.
func createPlan() shareModel {
	m := shareState()
	m.ID = types.StringUnknown()
	m.IP = types.StringUnknown()
	m.Metadata = metadata(map[string]string{"owner": "team"})
	m.AllMetadata = types.MapUnknown(types.StringType)
	m.Status = types.StringUnknown()
	m.AvailabilityZone = types.StringUnknown()
	m.ExportLocations = types.ListUnknown(types.ObjectType{AttrTypes: exportLocationTypes})
	m.Timeouts = shortTimeouts("2s")

	return m
}

func emptyState(t *testing.T, r *shareResource) tfsdk.State {
	t.Helper()
	state := shareStateOf(t, r, shareState())
	state.Raw = tftypes.NewValue(state.Raw.Type(), nil)

	return state
}

func readShareModel(t *testing.T, state tfsdk.State) shareModel {
	t.Helper()
	var m shareModel
	require.False(t, state.Get(context.Background(), &m).HasError())

	return m
}

func create(t *testing.T, r *shareResource, plan shareModel) (tfsdk.State, diag.Diagnostics) {
	t.Helper()
	resp := resource.CreateResponse{State: emptyState(t, r)}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(shareStateOf(t, r, plan))}, &resp)

	return resp.State, resp.Diagnostics
}

func read(t *testing.T, r *shareResource, prior shareModel) (tfsdk.State, diag.Diagnostics) {
	t.Helper()
	state := shareStateOf(t, r, prior)
	resp := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)

	return resp.State, resp.Diagnostics
}

func update(t *testing.T, r *shareResource, prior, plan shareModel) (shareModel, diag.Diagnostics) {
	t.Helper()
	planned := tfsdk.Plan(shareStateOf(t, r, plan))
	resp := resource.UpdateResponse{State: tfsdk.State(planned)}
	r.Update(context.Background(), resource.UpdateRequest{State: shareStateOf(t, r, prior), Plan: planned}, &resp)

	return readShareModel(t, resp.State), resp.Diagnostics
}

func destroy(t *testing.T, r *shareResource, prior shareModel) diag.Diagnostics {
	t.Helper()
	state := shareStateOf(t, r, prior)
	resp := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)

	return resp.Diagnostics
}

func TestShareCreate(t *testing.T) {
	t.Run("sends an empty access list and waits for available", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.statuses = []string{"creating", "available"}
		r := fake.resource()

		state, diags := create(t, r, createPlan())

		require.False(t, diags.HasError(), "%v", diags)
		var sent struct {
			Metadata map[string]string `json:"metadata"`
		}
		require.NoError(t, json.Unmarshal([]byte(fake.bodies["POST /shares"]), &sent))
		token := sent.Metadata[createTokenKey]
		require.NotEmpty(t, token)
		assert.JSONEq(t, `{
			"share_proto": "NFS", "size": 50, "share_type": "basic.ru-9a", "name": "data",
			"metadata": {"owner": "team", "`+createTokenKey+`": "`+token+`"},
			"neutron_net_id": "`+networkID+`", "neutron_subnet_id": "`+subnetID+`",
			"access_list": []
		}`, fake.bodies["POST /shares"])
		got := readShareModel(t, state)
		assert.Equal(t, shareID, got.ID.ValueString())
		assert.Equal(t, "available", got.Status.ValueString())
		assert.Equal(t, "10.0.0.5", got.IP.ValueString())
		assert.Equal(t, networkID, got.NetworkID.ValueString())
		assert.Equal(t, subnetID, got.SubnetID.ValueString())
		assert.Equal(t, "basic.ru-9a", got.ShareType.ValueString())
		assert.Equal(t, "ru-9a", got.AvailabilityZone.ValueString())
		assert.Len(t, got.ExportLocations.Elements(), 1)
		assert.Equal(t, metadata(map[string]string{"owner": "team"}), got.Metadata)
		assert.Contains(t, got.AllMetadata.Elements(), deleteNetworkKey)
	})

	t.Run("the creation token stays on the share and is hidden from all_metadata", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.statuses = []string{"creating", "available"}
		r := fake.resource()

		state, diags := create(t, r, createPlan())

		require.False(t, diags.HasError(), "%v", diags)
		assert.Contains(t, fake.share["metadata"], createTokenKey)
		got := readShareModel(t, state)
		assert.NotContains(t, got.AllMetadata.Elements(), createTokenKey)
		assert.Contains(t, got.AllMetadata.Elements(), deleteNetworkKey)
	})

	t.Run("a reserved key from a map unknown at validation sends no request", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()
		plan := createPlan()
		plan.Metadata = metadata(map[string]string{deleteNetworkKey: "disabled"})

		state, diags := create(t, r, plan)

		require.True(t, diags.HasError())
		assert.Equal(t, "Reserved metadata key", diags.Errors()[0].Summary())
		assert.Empty(t, fake.writes())
		assert.True(t, state.Raw.IsNull())
	})

	t.Run("a transient read failure does not end the wait", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.statuses = []string{"creating", "available"}
		fake.failOnce["GET /shares/"+shareID] = []answer{unavailable503}
		r := fake.resource()

		state, diags := create(t, r, createPlan())

		require.False(t, diags.HasError(), "%v", diags)
		got := readShareModel(t, state)
		assert.Equal(t, shareID, got.ID.ValueString())
		assert.Equal(t, "available", got.Status.ValueString())
		assert.Equal(t, 1, fake.count("POST /shares"))
	})

	for name, statuses := range map[string][]string{
		"error status": {"creating", "error"},
		"timeout":      {"creating"},
	} {
		t.Run(name+" keeps the UUID", func(t *testing.T) {
			fake := newFakeAPI(t)
			fake.statuses = statuses
			r := fake.resource()
			plan := createPlan()
			plan.Timeouts = shortTimeouts("200ms")

			state, diags := create(t, r, plan)

			require.True(t, diags.HasError())
			assert.Equal(t, "Wait for created file storage", diags.Errors()[0].Summary())
			assert.Equal(t, shareID, readShareModel(t, state).ID.ValueString())
			assert.Equal(t, 1, fake.count("POST /shares"))
		})
	}

	t.Run("an error status reports the API message", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.statuses = []string{"creating", "error"}
		fake.messages = []map[string]any{{
			"id": "message-1", "resource_type": "SHARE", "resource_id": shareID, "message_level": "ERROR",
			"user_message": "create: Could not find an existing share server.",
		}}
		r := fake.resource()

		_, diags := create(t, r, createPlan())

		require.True(t, diags.HasError())
		assert.Equal(t,
			"file storage "+shareID+" entered status error: create: Could not find an existing share server.",
			diags.Errors()[0].Detail())
		assert.Equal(t, shareID, fake.requestQuery("GET /messages", "resource_id"))
	})

	t.Run("a structured rejection leaves no state", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.fail["POST /shares"] = answer{http.StatusBadRequest, apiFault400}
		r := fake.resource()

		state, diags := create(t, r, createPlan())

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "IP 10.0.0.5 is already used.")
		assert.True(t, state.Raw.IsNull())
	})

	t.Run("an unsupported microversion names the pool", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.fail["POST /shares"] = answer{
			http.StatusNotAcceptable,
			`{"computeFault": {"code": 406, "message": "Version 2.75 is not supported by the API."}}`,
		}
		r := fake.resource()

		state, diags := create(t, r, createPlan())

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "pool ru-9")
		assert.Contains(t, diags.Errors()[0].Detail(), "microversion 2.75")
		assert.True(t, state.Raw.IsNull())
	})
}

func TestShareRead(t *testing.T) {
	t.Run("refreshes values changed outside Terraform", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share["name"] = "renamed"
		fake.share["size"] = 70
		r := fake.resource()

		state, diags := read(t, r, shareState())

		require.False(t, diags.HasError(), "%v", diags)
		got := readShareModel(t, state)
		assert.Equal(t, "renamed", got.Name.ValueString())
		assert.Equal(t, int64(70), got.Size.ValueInt64())
	})

	for name, tc := range map[string]struct {
		remote map[string]any
		want   map[string]string
	}{
		"a changed key shows the remote value": {map[string]any{"owner": "other"}, map[string]string{"owner": "other"}},
		"a removed key is dropped":             {map[string]any{}, map[string]string{}},
	} {
		t.Run(name, func(t *testing.T) {
			fake := newFakeAPI(t)
			remote := map[string]any{deleteNetworkKey: "enabled", "neutron_subnet_id": subnetID}
			maps.Copy(remote, tc.remote)
			fake.share["metadata"] = remote
			r := fake.resource()
			prior := shareState()
			prior.Metadata = metadata(map[string]string{"owner": "team"})

			state, diags := read(t, r, prior)

			require.False(t, diags.HasError(), "%v", diags)
			got := readShareModel(t, state)
			assert.Equal(t, metadata(tc.want), got.Metadata)
			assert.Contains(t, got.AllMetadata.Elements(), deleteNetworkKey)
		})
	}

	t.Run("service keys stay out of metadata", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()

		state, diags := read(t, r, shareState())

		require.False(t, diags.HasError(), "%v", diags)
		got := readShareModel(t, state)
		assert.True(t, got.Metadata.IsNull())
		assert.Equal(t, metadata(map[string]string{deleteNetworkKey: "enabled", "neutron_subnet_id": subnetID}), got.AllMetadata)
	})

	for name, tc := range map[string]struct {
		proto, path, ip string
	}{
		"NFS":            {"NFS", "10.0.0.7:/shares/share-1", "10.0.0.7"},
		"CIFS":           {"CIFS", `\\10.0.0.8\share-1`, "10.0.0.8"},
		"lowercase CIFS": {"cifs", `\\10.0.0.9\share-1`, "10.0.0.9"},
	} {
		t.Run("ip from the "+name+" path", func(t *testing.T) {
			fake := newFakeAPI(t)
			fake.share["share_proto"] = tc.proto
			fake.paths = []map[string]any{
				{"path": "10.9.9.9:/other", "preferred": false},
				{"path": tc.path, "preferred": true},
			}
			r := fake.resource()

			state, diags := read(t, r, shareState())

			require.False(t, diags.HasError(), "%v", diags)
			assert.Equal(t, tc.ip, readShareModel(t, state).IP.ValueString())
			assert.Empty(t, diags.Warnings())
		})
	}

	t.Run("a path without IPv4 keeps the address", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.paths = []map[string]any{{"path": "storage.local:/shares/share-1", "preferred": true}}
		r := fake.resource()

		state, diags := read(t, r, shareState())

		require.False(t, diags.HasError(), "%v", diags)
		require.Len(t, diags.Warnings(), 1)
		assert.Equal(t, "10.0.0.5", readShareModel(t, state).IP.ValueString())
	})

	t.Run("not found removes the share", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share = nil
		r := fake.resource()

		state, diags := read(t, r, shareState())

		require.False(t, diags.HasError(), "%v", diags)
		assert.True(t, state.Raw.IsNull())
	})

	for name, failure := range map[string]answer{
		"forbidden":    {http.StatusForbidden, `{"forbidden": {"code": 403, "message": "Policy does not allow."}}`},
		"server error": {http.StatusInternalServerError, `{"computeFault": {"code": 500, "message": "boom"}}`},
		"microversion": {http.StatusNotAcceptable, `{"computeFault": {"code": 406, "message": "Version 2.75 is not supported by the API."}}`},
	} {
		t.Run(name+" keeps the state", func(t *testing.T) {
			fake := newFakeAPI(t)
			fake.fail["GET /shares/"+shareID] = failure
			r := fake.resource()

			state, diags := read(t, r, shareState())

			require.True(t, diags.HasError())
			assert.Equal(t, shareID, readShareModel(t, state).ID.ValueString())
			if name == "microversion" {
				assert.Contains(t, diags.Errors()[0].Detail(), "pool ru-9")
			}
		})
	}

	t.Run("a missing network is an error, not a missing share", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.subnets = 0
		r := fake.resource()

		state, diags := read(t, r, shareState())

		require.True(t, diags.HasError())
		assert.False(t, state.Raw.IsNull())
	})
}

func TestShareImport(t *testing.T) {
	t.Run("restores every attribute and leaves metadata empty", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()
		state := emptyState(t, r)
		imported := resource.ImportStateResponse{State: state}

		r.ImportState(context.Background(), resource.ImportStateRequest{ID: shareID}, &imported)
		require.False(t, imported.Diagnostics.HasError(), "%v", imported.Diagnostics)
		resp := resource.ReadResponse{State: imported.State}
		r.Read(context.Background(), resource.ReadRequest{State: imported.State}, &resp)

		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		got := readShareModel(t, resp.State)
		assert.Equal(t, "project", got.ProjectID.ValueString())
		assert.Equal(t, "ru-9", got.Region.ValueString())
		assert.Equal(t, "10.0.0.5", got.IP.ValueString())
		assert.Equal(t, networkID, got.NetworkID.ValueString())
		assert.Equal(t, "NFS", got.ShareProto.ValueString())
		assert.True(t, got.Metadata.IsNull())
		assert.NotEmpty(t, got.AllMetadata.Elements())
	})

	t.Run("a path without IPv4 leaves ip empty", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.paths = []map[string]any{{"path": "storage.local:/shares/share-1", "preferred": true}}
		r := fake.resource()
		prior := shareState()
		prior.IP = types.StringNull()

		state, diags := read(t, r, prior)

		require.False(t, diags.HasError(), "%v", diags)
		require.Len(t, diags.Warnings(), 1)
		assert.Empty(t, readShareModel(t, state).IP.ValueString())
	})

	t.Run("a missing share is not found", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share = nil
		r := fake.resource()
		prior := shareState()

		state, diags := read(t, r, prior)

		require.False(t, diags.HasError())
		assert.True(t, state.Raw.IsNull(), "Terraform reports a removed import as a missing object")
	})

	for name, tc := range map[string]struct {
		id, project, region, detail string
	}{
		"no project": {shareID, "", "ru-9", "INFRA_PROJECT_ID"},
		"no region":  {shareID, "project", "", "INFRA_REGION"},
		"invalid ID": {"share", "project", "ru-9", "UUID"},
	} {
		t.Run(name, func(t *testing.T) {
			r := &shareResource{connection: connection{config: testConfig{project: tc.project, region: tc.region}}}
			resp := resource.ImportStateResponse{State: emptyState(t, r)}

			r.ImportState(context.Background(), resource.ImportStateRequest{ID: tc.id}, &resp)

			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.detail)
			assert.True(t, resp.State.Raw.IsNull())
		})
	}
}

func TestShareUpdate(t *testing.T) {
	t.Run("removing a key deletes only that key", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share["metadata"].(map[string]any)["owner"] = "team"
		fake.share["metadata"].(map[string]any)["env"] = "prod"
		r := fake.resource()
		prior := shareState()
		prior.Metadata = metadata(map[string]string{"owner": "team", "env": "prod"})
		prior.AllMetadata = metadata(map[string]string{"owner": "team", "env": "prod", deleteNetworkKey: "enabled"})
		plan := prior
		plan.Metadata = metadata(map[string]string{"owner": "team"})

		got, diags := update(t, r, prior, plan)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, []string{"DELETE /shares/" + shareID + "/metadata/env"}, fake.writes())
		assert.Equal(t, metadata(map[string]string{"owner": "team"}), got.Metadata)
		assert.Contains(t, got.AllMetadata.Elements(), deleteNetworkKey)
	})

	t.Run("a failed key deletion keeps the key managed", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share["metadata"].(map[string]any)["owner"] = "team"
		fake.share["metadata"].(map[string]any)["env"] = "prod"
		fake.fail["DELETE /shares/"+shareID+"/metadata/env"] = answer{
			http.StatusForbidden,
			`{"forbidden": {"code": 403, "message": "Policy doesn't allow share:delete_share_metadata to be performed."}}`,
		}
		r := fake.resource()
		prior := shareState()
		prior.Metadata = metadata(map[string]string{"owner": "team", "env": "prod"})
		prior.AllMetadata = metadata(map[string]string{"owner": "team", "env": "prod", deleteNetworkKey: "enabled"})
		plan := prior
		plan.Metadata = metadata(map[string]string{"owner": "team"})
		plan.Size = types.Int64Value(60)

		got, diags := update(t, r, prior, plan)

		require.True(t, diags.HasError())
		// env is still on the share, so the next plan deletes it again.
		assert.Equal(t, metadata(map[string]string{"owner": "team", "env": "prod"}), got.Metadata)
		assert.Equal(t, int64(50), got.Size.ValueInt64())
		assert.Zero(t, fake.count("POST /shares/"+shareID+"/action"))
	})

	t.Run("a reserved key from a map unknown at validation keeps the prior state", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()
		prior := shareState()
		plan := prior
		plan.Name = types.StringValue("renamed")
		plan.Metadata = metadata(map[string]string{deleteNetworkKey: "disabled"})

		got, diags := update(t, r, prior, plan)

		require.True(t, diags.HasError())
		assert.Equal(t, "Reserved metadata key", diags.Errors()[0].Summary())
		assert.Empty(t, fake.writes())
		assert.Equal(t, prior.Name, got.Name)
		assert.Equal(t, prior.Metadata, got.Metadata)
	})

	t.Run("changed values are merged", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()
		prior := shareState()
		plan := prior
		plan.Name = types.StringValue("renamed")
		plan.Metadata = metadata(map[string]string{"owner": "team"})

		got, diags := update(t, r, prior, plan)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, []string{"PUT /shares/" + shareID, "POST /shares/" + shareID + "/metadata"}, fake.writes())
		assert.JSONEq(t, `{"metadata": {"owner": "team"}}`, fake.bodies["POST /shares/"+shareID+"/metadata"])
		assert.Equal(t, "renamed", got.Name.ValueString())
		assert.Contains(t, got.AllMetadata.Elements(), deleteNetworkKey)
	})

	t.Run("extension waits for available", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()
		prior := shareState()
		prior.Timeouts = shortTimeouts("2s")
		plan := prior
		plan.Size = types.Int64Value(60)
		fake.statuses = []string{"extending", "extending", "available"}

		got, diags := update(t, r, prior, plan)

		require.False(t, diags.HasError(), "%v", diags)
		assert.JSONEq(t, `{"extend": {"new_size": 60}}`, fake.bodies["POST /shares/"+shareID+"/action"])
		assert.Equal(t, int64(60), got.Size.ValueInt64())
		assert.Equal(t, "available", got.Status.ValueString())
	})

	t.Run("an available read at the old size reports the API message", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.statuses = []string{"rejected"}
		fake.messages = []map[string]any{{
			"id": "message-1", "resource_type": "SHARE", "resource_id": shareID, "message_level": "ERROR",
			"user_message": "extend: No storage could be allocated for this share request.",
		}}
		r := fake.resource()
		prior := shareState()
		prior.Timeouts = shortTimeouts("2s")
		plan := prior
		plan.Size = types.Int64Value(60)

		got, diags := update(t, r, prior, plan)

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(),
			"file storage "+shareID+" entered status available at size 50: "+
				"extend: No storage could be allocated for this share request.")
		assert.Equal(t, int64(50), got.Size.ValueInt64())
	})

	t.Run("a failed extension keeps the reported size", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.fail["POST /shares/"+shareID+"/action"] = answer{
			http.StatusForbidden,
			`{"forbidden": {"code": 403, "message": "Requested share exceeds allowed project quota."}}`,
		}
		r := fake.resource()
		prior := shareState()
		plan := prior
		plan.Size = types.Int64Value(60)

		got, diags := update(t, r, prior, plan)

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "quota")
		assert.Equal(t, int64(50), got.Size.ValueInt64())
	})

	t.Run("extending_error ends the wait", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.statuses = []string{"extending", "extending_error"}
		r := fake.resource()
		prior := shareState()
		prior.Timeouts = shortTimeouts("2s")
		plan := prior
		plan.Size = types.Int64Value(60)

		got, diags := update(t, r, prior, plan)

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "extending_error")
		assert.Equal(t, "extending_error", got.Status.ValueString())
	})
}

func TestShareDelete(t *testing.T) {
	t.Run("waits for the share and its network", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.statuses = []string{"available", "deleting", "deleted"}
		r := fake.resource()
		prior := shareState()
		prior.Timeouts = shortTimeouts("2s")

		diags := destroy(t, r, prior)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Empty(t, diags.Warnings())
		assert.Equal(t, []string{"DELETE /shares/" + shareID}, fake.writes())
		assert.GreaterOrEqual(t, fake.count("GET /share-networks/"+networkShareID), 2)
	})

	t.Run("transient read failures do not count as deleted", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.statuses = []string{"available", "deleting", "deleted"}
		// The reads before DELETE succeed; the first read of each wait fails.
		fake.failOnce["GET /shares/"+shareID] = []answer{{}, unavailable503}
		fake.failOnce["GET /share-networks/"+networkShareID] = []answer{{}, unavailable503}
		r := fake.resource()
		prior := shareState()
		prior.Timeouts = shortTimeouts("2s")

		diags := destroy(t, r, prior)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Nil(t, fake.share)
		assert.GreaterOrEqual(t, fake.count("GET /share-networks/"+networkShareID), 3)
	})

	t.Run("a network with other subnets is not waited for", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.subnets = 2
		fake.statuses = []string{"available", "deleted"}
		r := fake.resource()

		diags := destroy(t, r, shareState())

		require.False(t, diags.HasError(), "%v", diags)
		assert.Empty(t, diags.Warnings())
		assert.Equal(t, 1, fake.count("GET /share-networks/"+networkShareID))
	})

	t.Run("without the flag the network is not read", func(t *testing.T) {
		fake := newFakeAPI(t)
		delete(fake.share["metadata"].(map[string]any), deleteNetworkKey)
		fake.statuses = []string{"available", "deleted"}
		r := fake.resource()

		diags := destroy(t, r, shareState())

		require.False(t, diags.HasError(), "%v", diags)
		assert.Zero(t, fake.count("GET /share-networks/"+networkShareID))
	})

	t.Run("a missing share is deleted", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share = nil
		r := fake.resource()

		diags := destroy(t, r, shareState())

		require.False(t, diags.HasError(), "%v", diags)
		assert.Empty(t, fake.writes())
	})

	t.Run("error_deleting fails", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.statuses = []string{"available", "deleting", "error_deleting"}
		r := fake.resource()
		prior := shareState()
		prior.Timeouts = shortTimeouts("2s")

		diags := destroy(t, r, prior)

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "error_deleting")
	})

	t.Run("a refusal by status fails", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.fail["DELETE /shares/"+shareID] = answer{
			http.StatusForbidden,
			`{"forbidden": {"code": 403, "message": "Invalid share: Share status must be one of available, error."}}`,
		}
		r := fake.resource()

		diags := destroy(t, r, shareState())

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "Share status must be")
	})
}
