package images

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func apiMember(status string) map[string]any {
	return map[string]any{
		"image_id": imageID, "member_id": memberID, "status": status,
		"created_at": "2026-10-09T10:00:00Z", "updated_at": "2026-10-09T10:00:00Z",
	}
}

func accessState(t *testing.T, r resource.Resource, model any) tfsdk.State {
	t.Helper()
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	state := tfsdk.State{Schema: s.Schema}
	require.False(t, state.Set(context.Background(), model).HasError())

	return state
}

func plannedAccess() accessModel {
	m := accessModel{ProjectID: types.StringValue(projectID), Region: types.StringValue("ru-6")}
	m.ID, m.Status = types.StringUnknown(), types.StringUnknown()
	m.CreatedAt, m.UpdatedAt = types.StringUnknown(), types.StringUnknown()
	m.ImageID, m.MemberID = types.StringValue(imageID), types.StringValue(memberID)

	return m
}

func TestGrantAccess(t *testing.T) {
	for name, tc := range map[string]struct {
		createStatus int
		err          string
	}{
		"granted":        {http.StatusOK, ""},
		"already exists": {http.StatusConflict, "Import the access with the ID " + imageID + "/" + memberID + " to manage it."},
		// A lost answer is resolved by reading the membership.
		"answer lost": {http.StatusBadGateway, ""},
	} {
		t.Run(name, func(t *testing.T) {
			r := &accessResource{testConnection(t, func(w http.ResponseWriter, req *http.Request) {
				switch req.Method {
				case http.MethodPost:
					writeJSON(t, w, tc.createStatus, apiMember("pending"))
				case http.MethodGet:
					writeJSON(t, w, http.StatusOK, apiMember("pending"))
				}
			})}
			plan := plannedAccess()
			state := accessState(t, r, &plan)
			resp := resource.CreateResponse{State: tfsdk.State{
				Schema: state.Schema, Raw: tftypes.NewValue(state.Raw.Type(), nil),
			}}
			r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(state)}, &resp)
			if tc.err != "" {
				require.True(t, resp.Diagnostics.HasError())
				assert.Equal(t, tc.err, resp.Diagnostics.Errors()[0].Detail())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var created accessModel
			require.False(t, resp.State.Get(context.Background(), &created).HasError())
			assert.Equal(t, imageID+"/"+memberID, created.ID.ValueString())
			assert.Equal(t, "pending", created.Status.ValueString())
			assert.Equal(t, "2026-10-09T10:00:00Z", created.CreatedAt.ValueString())
		})
	}
}

func TestRevokeAccess(t *testing.T) {
	for name, tc := range map[string]struct {
		status int
		// detail is a part of the error detail; empty means success.
		detail string
	}{
		"revoked":         {http.StatusNoContent, ""},
		"already revoked": {http.StatusNotFound, ""},
		// The API rejects the request while the image is not shared.
		"image not shared": {http.StatusForbidden, "set visibility = \"shared\" on the image"},
		"server error":     {http.StatusInternalServerError, "status 500"},
	} {
		t.Run(name, func(t *testing.T) {
			var requested string
			r := &accessResource{testConnection(t, func(w http.ResponseWriter, req *http.Request) {
				requested = req.Method + " " + req.URL.Path
				if tc.status == http.StatusNoContent {
					w.WriteHeader(tc.status)
					return
				}
				writeJSON(t, w, tc.status, map[string]any{"message": "error"})
			})}
			prior := plannedAccess()
			prior.ID, prior.Status = types.StringValue(imageID+"/"+memberID), types.StringValue("pending")
			prior.CreatedAt, prior.UpdatedAt = types.StringValue("2026-10-09T10:00:00Z"), types.StringValue("2026-10-09T10:00:00Z")
			var resp resource.DeleteResponse
			r.Delete(context.Background(), resource.DeleteRequest{State: accessState(t, r, &prior)}, &resp)

			assert.Equal(t, "DELETE /v2/images/"+imageID+"/members/"+memberID, requested)
			if tc.detail == "" {
				require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
				return
			}
			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.detail)
		})
	}
}

// TestReadAccess covers both access resources: they read the same membership.
func TestReadAccess(t *testing.T) {
	for name, tc := range map[string]struct {
		status int
		gone   bool
		err    bool
	}{
		"found": {http.StatusOK, false, false},
		// The API also answers 404 while the image is not shared.
		"not found":    {http.StatusNotFound, true, false},
		"server error": {http.StatusInternalServerError, false, true},
	} {
		t.Run(name, func(t *testing.T) {
			conn := testConnection(t, func(w http.ResponseWriter, _ *http.Request) {
				reply := apiMember("accepted")
				reply["updated_at"] = "2026-10-10T10:00:00Z"
				writeJSON(t, w, tc.status, reply)
			})
			prior := membership{
				ID: types.StringValue(imageID + "/" + memberID), ImageID: types.StringValue(imageID),
				MemberID: types.StringValue(memberID), Status: types.StringValue("pending"),
				CreatedAt: types.StringValue("2026-10-09T10:00:00Z"), UpdatedAt: types.StringValue("2026-10-09T10:00:00Z"),
			}
			project, region := types.StringValue(memberID), types.StringValue("ru-6")
			for _, r := range []struct {
				resource resource.Resource
				model    any
			}{
				{&accessResource{conn}, &accessModel{ProjectID: project, Region: region, membership: prior}},
				{&accessAcceptResource{conn}, &accessAcceptModel{ProjectID: project, Region: region, membership: prior}},
			} {
				resp := resource.ReadResponse{State: accessState(t, r.resource, r.model)}
				r.resource.Read(context.Background(), resource.ReadRequest{State: accessState(t, r.resource, r.model)}, &resp)
				assert.Equal(t, tc.gone, resp.State.Raw.IsNull())
				assert.Equal(t, tc.err, resp.Diagnostics.HasError())
				if tc.status == http.StatusOK {
					for attribute, want := range map[string]string{"status": "accepted", "updated_at": "2026-10-10T10:00:00Z"} {
						var got types.String
						require.False(t, resp.State.GetAttribute(context.Background(), path.Root(attribute), &got).HasError())
						assert.Equal(t, want, got.ValueString(), attribute)
					}
				}
			}
		})
	}
}

func TestAcceptOwnAccessByDefault(t *testing.T) {
	var requested string
	r := &accessAcceptResource{testConnection(t, func(w http.ResponseWriter, req *http.Request) {
		requested = req.Method + " " + req.URL.Path
		writeJSON(t, w, http.StatusOK, map[string]any{
			"image_id": imageID, "member_id": projectID, "status": "accepted",
			"created_at": "2026-10-09T10:00:00Z", "updated_at": "2026-10-09T10:00:00Z",
		})
	})}
	plan := accessAcceptModel{ProjectID: types.StringValue(projectID), Region: types.StringValue("ru-6")}
	plan.ImageID, plan.Status = types.StringValue(imageID), types.StringValue("accepted")
	plan.ID, plan.MemberID = types.StringUnknown(), types.StringUnknown()
	plan.CreatedAt, plan.UpdatedAt = types.StringUnknown(), types.StringUnknown()
	state := accessState(t, r, &plan)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: state.Schema, Raw: tftypes.NewValue(state.Raw.Type(), nil)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(state)}, &resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	// An omitted member_id is the project of the resource.
	assert.Equal(t, "PUT /v2/images/"+imageID+"/members/"+projectID, requested)
	var created accessAcceptModel
	require.False(t, resp.State.Get(context.Background(), &created).HasError())
	assert.Equal(t, imageID+"/"+projectID, created.ID.ValueString())
	assert.Equal(t, projectID, created.MemberID.ValueString())
}

func TestDestroyRejectsAccess(t *testing.T) {
	for name, tc := range map[string]struct {
		status int
		err    bool
	}{
		"rejected":     {http.StatusOK, false},
		"access gone":  {http.StatusNotFound, false},
		"server error": {http.StatusInternalServerError, true},
	} {
		t.Run(name, func(t *testing.T) {
			var sent map[string]string
			r := &accessAcceptResource{testConnection(t, func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, http.MethodPut, req.Method)
				assert.NoError(t, json.NewDecoder(req.Body).Decode(&sent))
				writeJSON(t, w, tc.status, apiMember("rejected"))
			})}
			prior := accessAcceptModel{ProjectID: types.StringValue(memberID), Region: types.StringValue("ru-6")}
			prior.ID, prior.ImageID = types.StringValue(imageID+"/"+memberID), types.StringValue(imageID)
			prior.MemberID, prior.Status = types.StringValue(memberID), types.StringValue("accepted")
			prior.CreatedAt, prior.UpdatedAt = types.StringValue("2026-10-09T10:00:00Z"), types.StringValue("2026-10-09T10:00:00Z")
			var resp resource.DeleteResponse
			r.Delete(context.Background(), resource.DeleteRequest{State: accessState(t, r, &prior)}, &resp)
			assert.Equal(t, tc.err, resp.Diagnostics.HasError(), resp.Diagnostics)
			assert.Equal(t, map[string]string{"status": "rejected"}, sent)
		})
	}
}

func TestAcceptOnlyOwnAccess(t *testing.T) {
	r := &accessAcceptResource{}
	for name, tc := range map[string]struct {
		member types.String
		valid  bool
	}{
		"omitted": {types.StringNull(), true},
		// A member_id that is not known yet is checked during apply.
		"not known yet": {types.StringUnknown(), true},
		"own":           {types.StringValue(projectID), true},
		"other project": {types.StringValue(memberID), false},
	} {
		t.Run(name, func(t *testing.T) {
			config := accessAcceptModel{ProjectID: types.StringValue(projectID), Region: types.StringValue("ru-6")}
			config.ImageID, config.MemberID, config.Status = types.StringValue(imageID), tc.member, types.StringValue("accepted")
			config.ID, config.CreatedAt, config.UpdatedAt = types.StringNull(), types.StringNull(), types.StringNull()
			var resp resource.ValidateConfigResponse
			r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{
				Config: tfsdk.Config(accessState(t, r, &config)),
			}, &resp)
			assert.Equal(t, tc.valid, !resp.Diagnostics.HasError())
		})
	}
}

// TestOmittedMemberIDKeepsAcceptance runs the member_id plan modifiers in
// schema order, as Framework does for any change of the resource, for
// example, of status: an omitted member_id is unknown in the plan and must
// not replace the resource.
func TestOmittedMemberIDKeepsAcceptance(t *testing.T) {
	ctx := context.Background()
	r := &accessAcceptResource{}
	var s resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &s)
	prior := accessAcceptModel{ProjectID: types.StringValue(projectID), Region: types.StringValue("ru-6")}
	prior.ID, prior.ImageID = types.StringValue(imageID+"/"+projectID), types.StringValue(imageID)
	prior.MemberID, prior.Status = types.StringValue(projectID), types.StringValue("accepted")
	prior.CreatedAt, prior.UpdatedAt = types.StringValue("2026-10-09T10:00:00Z"), types.StringValue("2026-10-09T10:00:00Z")
	state := accessState(t, r, &prior)

	req := planmodifier.StringRequest{
		Path: path.Root("member_id"), State: state, Plan: tfsdk.Plan(state),
		ConfigValue: types.StringNull(), StateValue: prior.MemberID, PlanValue: types.StringUnknown(),
	}
	resp := planmodifier.StringResponse{PlanValue: req.PlanValue}
	for _, modifier := range s.Schema.Attributes["member_id"].(schema.StringAttribute).PlanModifiers {
		req.PlanValue = resp.PlanValue
		modifier.PlanModifyString(ctx, req, &resp)
	}
	assert.False(t, resp.RequiresReplace)
	assert.Equal(t, prior.MemberID, resp.PlanValue)
}

func TestParseMembershipID(t *testing.T) {
	for name, tc := range map[string]struct {
		id             string
		memberRequired bool
		member         string
		valid          bool
	}{
		"full":                 {imageID + "/" + memberID, true, memberID, true},
		"image only":           {imageID, false, "", true},
		"image only, required": {imageID, true, "", false},
		"empty member":         {imageID + "/", false, "", false},
		"member with slash":    {imageID + "/a/b", false, "", false},
		"uppercase image ID":   {"11111111-1111-4111-8111-11111111111A/" + memberID, true, "", false},
	} {
		parsedImage, parsedMember, err := parseMembershipID(tc.id, tc.memberRequired)
		assert.Equal(t, tc.valid, err == nil, name)
		if tc.valid {
			assert.Equal(t, imageID, parsedImage, name)
			assert.Equal(t, tc.member, parsedMember, name)
		}
	}
}

func TestImportAccessAcceptance(t *testing.T) {
	r := &accessAcceptResource{connection{config: testConfig{}}}
	// The full ID names the project that received access; the short one
	// takes the project of the provider.
	for id, project := range map[string]string{imageID + "/" + memberID: memberID, imageID: projectID} {
		var s resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &s)
		resp := resource.ImportStateResponse{State: tfsdk.State{
			Schema: s.Schema, Raw: tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil),
		}}
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
		var imported accessAcceptModel
		require.False(t, resp.State.Get(context.Background(), &imported).HasError())
		assert.Equal(t, imageID+"/"+project, imported.ID.ValueString(), id)
		assert.Equal(t, imageID, imported.ImageID.ValueString(), id)
		assert.Equal(t, project, imported.ProjectID.ValueString(), id)
		assert.Equal(t, project, imported.MemberID.ValueString(), id)
	}
}
