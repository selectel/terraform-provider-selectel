package images

import (
	"cmp"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/selectel/images-go/pkg/v1/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImagePatch(t *testing.T) {
	current := &image.View{
		Name: "image", Visibility: "private", Protected: true, Hidden: true, MinDisk: 3, MinRAM: 512,
		Tags:       []string{"b", "c"},
		Properties: map[string]string{"os_distro": "ubuntu", "kept": "old", "removed": "x", "hw_qemu_guest_agent": "yes"},
	}
	plan := planned()
	plan.Name = types.StringValue("renamed")
	plan.Visibility = types.StringValue("shared")
	plan.MinDiskGB = types.Int64Value(5)
	plan.Tags = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	plan.Properties = properties(map[string]string{"kept": "new", "added": "1", "a/b": "2"})
	// gone was managed before but is no longer on the image: removing it would fail.
	previouslyManaged := properties(map[string]string{"kept": "old", "removed": "x", "gone": "y"})

	ops, diags := imagePatch(context.Background(), previouslyManaged, plan, current)
	require.False(t, diags.HasError())
	// The property that the API added produces no operation. TestUpdateImage
	// covers tags in another order, and TestCreateTransfersData an unknown
	// min_disk.
	assert.Equal(t, []image.PatchOp{
		{Op: image.OpReplace, Path: "/name", Value: "renamed"},
		{Op: image.OpReplace, Path: "/visibility", Value: "shared"},
		{Op: image.OpReplace, Path: "/protected", Value: false},
		{Op: image.OpReplace, Path: "/os_hidden", Value: false},
		{Op: image.OpReplace, Path: "/min_ram", Value: int64(0)},
		{Op: image.OpReplace, Path: "/min_disk", Value: int64(5)},
		{Op: image.OpReplace, Path: "/tags", Value: []string{"a", "b"}},
		{Op: image.OpAdd, Path: "/a~1b", Value: "2"},
		{Op: image.OpAdd, Path: "/added", Value: "1"},
		{Op: image.OpReplace, Path: "/kept", Value: "new"},
		{Op: image.OpRemove, Path: "/removed"},
	}, ops)
}

// fakeImageAPI is the part of the Image service that Create uses. The PUT
// handler stores the SHA-512 of the received data, as the service does.
type fakeImageAPI struct {
	t       *testing.T
	image   map[string]any
	hash    func(received []byte) string
	patches [][]image.PatchOp
	// hashAlgo is the checksum algorithm that the service reports; empty is sha512.
	hashAlgo string
	// importFailure is the message of the failed import task. Without it,
	// the import succeeds after one read that reports the image as importing.
	importFailure string
	// importRejected makes the import request fail with 409.
	importRejected bool
	// importHangs keeps the import request unanswered, as the service does
	// for some URLs.
	importHangs bool
	// stuck keeps the image in progress after the data transfer.
	stuck bool
	// unavailable is the number of image reads that fail with 503 first.
	unavailable int
	// createFails, uploadFails and patchFails make these requests fail.
	createFails, uploadFails, patchFails bool
}

func (f *fakeImageAPI) serve(w http.ResponseWriter, req *http.Request) {
	switch req.Method + " " + req.URL.Path {
	case "POST /v2/images":
		if f.createFails {
			writeJSON(f.t, w, http.StatusInternalServerError, map[string]any{"message": "error"})
			return
		}
		var opts map[string]any
		require.NoError(f.t, json.NewDecoder(req.Body).Decode(&opts))
		f.image = apiImage(opts)
		f.image["status"], f.image["min_disk"] = "queued", 0
		writeJSON(f.t, w, http.StatusCreated, f.image)
	case "PUT /v2/images/" + imageID + "/file":
		received, err := io.ReadAll(req.Body)
		require.NoError(f.t, err)
		if f.uploadFails {
			writeJSON(f.t, w, http.StatusInternalServerError, map[string]any{"message": "error"})
			return
		}
		f.image["status"], f.image["min_disk"] = "active", 1
		if f.stuck {
			f.image["status"] = "saving"
		}
		f.image["os_hash_algo"], f.image["os_hash_value"] = cmp.Or(f.hashAlgo, "sha512"), f.hash(received)
		w.WriteHeader(http.StatusNoContent)
	case "POST /v2/images/" + imageID + "/import":
		if f.importRejected {
			writeJSON(f.t, w, http.StatusConflict, map[string]any{"message": "conflict"})
			return
		}
		if f.importHangs {
			// The server notices the closed connection only after the body is read.
			_, err := io.Copy(io.Discard, req.Body)
			require.NoError(f.t, err)
			<-req.Context().Done()

			return
		}
		f.image["status"] = "importing"
		w.WriteHeader(http.StatusAccepted)
	case "GET /v2/images/" + imageID + "/tasks":
		// Only the newest import task tells the outcome of the download.
		tasks := []map[string]any{{"type": "api_image_import", "status": "success", "created_at": "2026-10-09T10:00:00Z"}}
		if f.importFailure != "" {
			tasks = []map[string]any{
				{"type": "api_image_import", "status": "failure", "message": "older_failure", "created_at": "2026-10-08T10:00:00Z"},
				{"type": "api_image_import", "status": "failure", "message": f.importFailure, "created_at": "2026-10-09T10:00:00Z"},
				{"type": "api_image_import", "status": "failure", "message": "oldest_failure", "created_at": "2026-10-07T10:00:00Z"},
				{"type": "location_import", "status": "failure", "message": "not_an_import", "created_at": "2026-10-09T11:00:00Z"},
			}
		}
		writeJSON(f.t, w, http.StatusOK, map[string]any{"tasks": tasks})
	case "GET /v2/images/" + imageID:
		if f.unavailable > 0 {
			f.unavailable--
			writeJSON(f.t, w, http.StatusServiceUnavailable, map[string]any{"message": "unavailable"})
			return
		}
		writeJSON(f.t, w, http.StatusOK, f.image)
		if f.image["status"] == "importing" && f.importFailure == "" && !f.stuck {
			f.image["status"], f.image["min_disk"] = "active", 1
		}
	case "PATCH /v2/images/" + imageID:
		if f.patchFails {
			writeJSON(f.t, w, http.StatusConflict, map[string]any{"message": "conflict"})
			return
		}
		var ops []image.PatchOp
		require.NoError(f.t, json.NewDecoder(req.Body).Decode(&ops))
		f.patches = append(f.patches, ops)
		for _, op := range ops {
			f.image[op.Path[1:]] = op.Value
		}
		writeJSON(f.t, w, http.StatusOK, f.image)
	default:
		f.t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func sha512Hex(received []byte) string {
	digest := sha512.Sum512(received)
	return hex.EncodeToString(digest[:])
}

func create(t *testing.T, r *imageResource, plan imageModel) resource.CreateResponse {
	t.Helper()
	state := imageState(t, r, plan)
	resp := resource.CreateResponse{State: tfsdk.State{
		Schema: state.Schema,
		Raw:    tftypes.NewValue(state.Schema.Type().TerraformType(context.Background()), nil),
	}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(state)}, &resp)

	return resp
}

// sourcePlan is the planned image with data from a local file or from a URL.
func sourcePlan(t *testing.T, upload bool) imageModel {
	t.Helper()
	plan := planned()
	if !upload {
		plan.ImageSourceURL = types.StringValue("https://example.com/disk.img")
		return plan
	}
	file := filepath.Join(t.TempDir(), "disk.raw")
	require.NoError(t, os.WriteFile(file, []byte("data"), 0o600))
	plan.LocalFilePath = types.StringValue(file)

	return plan
}

func TestCreateTransfersData(t *testing.T) {
	protect := &image.PatchOp{Op: image.OpReplace, Path: "/protected", Value: true}
	for name, tc := range map[string]struct {
		api       *fakeImageAPI
		upload    bool
		protected bool
		err       string
	}{
		// A failed read does not end the wait: the next one sees the image active.
		// The protection is enabled only after the data is in place.
		"uploaded":          {&fakeImageAPI{hash: sha512Hex, unavailable: 1}, true, true, ""},
		"downloaded":        {&fakeImageAPI{}, false, false, ""},
		"checksum mismatch": {&fakeImageAPI{hash: func([]byte) string { return "other" }}, true, true, "Verify uploaded image"},
		"other checksum":    {&fakeImageAPI{hash: sha512Hex, hashAlgo: "md5"}, true, true, "Verify uploaded image"},
	} {
		t.Run(name, func(t *testing.T) {
			tc.api.t = t
			r := testImageResource(t, tc.api.serve)
			plan := sourcePlan(t, tc.upload)
			plan.Protected = types.BoolValue(tc.protected)

			resp := create(t, r, plan)
			created := readImage(t, resp.State)
			assert.Equal(t, imageID, created.ID.ValueString())
			if tc.err != "" {
				require.True(t, resp.Diagnostics.HasError())
				assert.Equal(t, tc.err, resp.Diagnostics.Errors()[0].Summary())
				// The registered image stays in the state, so Terraform taints it.
				assert.Empty(t, tc.api.patches)

				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var patches [][]image.PatchOp
			if tc.protected {
				patches = [][]image.PatchOp{{*protect}}
			}
			assert.Equal(t, patches, tc.api.patches)
			assert.Equal(t, tc.protected, created.Protected.ValueBool())
			assert.Equal(t, int64(1), created.MinDiskGB.ValueInt64(), "min_disk derived by the API")
		})
	}
}

func TestCreateExplainsFailure(t *testing.T) {
	for name, tc := range map[string]struct {
		api    *fakeImageAPI
		upload bool
		// summary and cause are the error; hint is the advice that ends the
		// detail, if any.
		summary, cause, hint string
		registered           bool
	}{
		"no definite answer": {
			&fakeImageAPI{createFails: true}, false,
			"Create image", "status 500", fmt.Sprintf(uncertainCreateHint, "image"), false,
		},
		"upload failed":     {&fakeImageAPI{uploadFails: true}, true, "Transfer image data", "upload", "", true},
		"download rejected": {&fakeImageAPI{importRejected: true}, false, "Transfer image data", "status 409", "", true},
		"request not accepted": {
			&fakeImageAPI{importHangs: true}, false, "Transfer image data", "the request timed out", importTimeoutHint, true,
		},
		"download failed": {
			&fakeImageAPI{importFailure: "url_not_accessible"}, false,
			"Wait for active image", "could not download the image data: url_not_accessible", "", true,
		},
		"data not ready in time": {&fakeImageAPI{stuck: true}, false, "Wait for active image", "wait for image", untaintHint, true},
		"settings not applied":   {&fakeImageAPI{hash: sha512Hex, patchFails: true}, true, "Apply image settings", "conflict", "", true},
	} {
		t.Run(name, func(t *testing.T) {
			tc.api.t = t
			r := testImageResource(t, tc.api.serve)
			plan := sourcePlan(t, tc.upload)
			plan.Protected = types.BoolValue(true)

			resp := create(t, r, plan)
			require.True(t, resp.Diagnostics.HasError())
			failure := resp.Diagnostics.Errors()[0]
			assert.Equal(t, tc.summary, failure.Summary())
			assert.Contains(t, failure.Detail(), tc.cause)
			if tc.hint == "" {
				assert.NotContains(t, failure.Detail(), "\n\n")
			} else {
				assert.True(t, strings.HasSuffix(failure.Detail(), "\n\n"+tc.hint), failure.Detail())
			}
			// A registered image stays in the state, so Terraform taints it.
			assert.Equal(t, !tc.registered, resp.State.Raw.IsNull())
		})
	}
}

func TestUpdateImage(t *testing.T) {
	// The image was renamed outside of Terraform after the last refresh.
	remote := apiImage(map[string]any{"name": "renamed outside", "tags": []string{"b", "a"}})
	prior := planned()
	prior.ID = types.StringValue(imageID)
	prior.MinDiskGB = types.Int64Value(1)
	prior.Tags = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	prior.Properties = properties(map[string]string{"os_distro": "ubuntu"})
	for name, tc := range map[string]struct {
		change  func(*imageModel)
		reject  bool
		patches [][]image.PatchOp
		// name is the name in the state after the update.
		name string
	}{
		"applied": {
			change: func(m *imageModel) {
				m.Name, m.Tags, m.Properties = types.StringValue("new name"), types.SetNull(types.StringType), types.MapNull(types.StringType)
			},
			// Omitted tags clear the tags of the image.
			patches: [][]image.PatchOp{{
				{Op: image.OpReplace, Path: "/name", Value: "new name"},
				{Op: image.OpReplace, Path: "/tags", Value: []any{}},
				{Op: image.OpRemove, Path: "/os_distro"},
			}},
			name: "new name",
		},
		// Tags in another order and a known min_disk equal to the image's
		// produce no request.
		"only sources changed": {
			change: func(m *imageModel) {
				m.Name, m.LocalFilePath = types.StringValue("renamed outside"), types.StringValue("/tmp/a.raw")
			},
			name: "renamed outside",
		},
		// The Image service applies a patch as a whole, so the state is the
		// image read before the failed patch.
		"rejected": {
			change: func(m *imageModel) {
				m.Name, m.Properties = types.StringValue("new name"), properties(map[string]string{"os_distro": "debian"})
			},
			reject: true,
			name:   "renamed outside",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var patches [][]image.PatchOp
			r := testImageResource(t, func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodPatch {
					writeJSON(t, w, http.StatusOK, remote)
					return
				}
				if tc.reject {
					writeJSON(t, w, http.StatusConflict, map[string]any{"message": "conflict"})
					return
				}
				var ops []image.PatchOp
				require.NoError(t, json.NewDecoder(req.Body).Decode(&ops))
				patches = append(patches, ops)
				writeJSON(t, w, http.StatusOK, apiImage(map[string]any{"name": "new name"}))
			})
			plan := prior
			tc.change(&plan)

			resp := resource.UpdateResponse{State: imageState(t, r, prior)}
			r.Update(context.Background(), resource.UpdateRequest{
				Plan: tfsdk.Plan(imageState(t, r, plan)), State: imageState(t, r, prior),
			}, &resp)
			assert.Equal(t, tc.reject, resp.Diagnostics.HasError(), resp.Diagnostics)
			assert.Equal(t, tc.patches, patches)
			state := readImage(t, resp.State)
			assert.Equal(t, tc.name, state.Name.ValueString())
			assert.Equal(t, "2026-10-09T10:00:00Z", state.UpdatedAt.ValueString(), "observed after the update")
			want := plan
			if tc.reject {
				want = prior
			}
			assert.Equal(t, want.Properties, state.Properties)
			assert.Equal(t, want.LocalFilePath, state.LocalFilePath)
		})
	}
}

func TestDeleteImage(t *testing.T) {
	type reply struct {
		status int
		image  map[string]any
	}
	active, pendingDelete := reply{http.StatusOK, apiImage(nil)}, reply{http.StatusOK, apiImage(map[string]any{"status": "pending_delete"})}
	unavailable := reply{http.StatusServiceUnavailable, map[string]any{"message": "unavailable"}}
	for name, tc := range map[string]struct {
		// reads answers the image reads in order; the last answer repeats.
		reads []reply
		// deleteStatus answers the delete request; zero is no request.
		deleteStatus int
		err          string
	}{
		"deleted":      {[]reply{active, active, unavailable, pendingDelete}, http.StatusNoContent, ""},
		"already gone": {[]reply{{http.StatusNotFound, map[string]any{"message": "not found"}}}, 0, ""},
		// Another delete request removed the image after the read.
		"deleted concurrently": {[]reply{active, pendingDelete}, http.StatusNotFound, ""},
		"protected":            {[]reply{{http.StatusOK, apiImage(map[string]any{"protected": true})}}, 0, "Image is protected"},
		"delete rejected":      {[]reply{active}, http.StatusConflict, "Delete image"},
		"not deleted in time":  {[]reply{active}, http.StatusNoContent, "Wait for deleted image"},
	} {
		t.Run(name, func(t *testing.T) {
			reads, deleted := 0, false
			r := testImageResource(t, func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodDelete {
					deleted = true
					if tc.deleteStatus == http.StatusNoContent {
						w.WriteHeader(tc.deleteStatus)
						return
					}
					writeJSON(t, w, tc.deleteStatus, map[string]any{"message": "delete"})

					return
				}
				answer := tc.reads[min(reads, len(tc.reads)-1)]
				reads++
				writeJSON(t, w, answer.status, answer.image)
			})
			prior := planned()
			prior.ID = types.StringValue(imageID)
			resp := resource.DeleteResponse{State: imageState(t, r, prior)}
			r.Delete(context.Background(), resource.DeleteRequest{State: imageState(t, r, prior)}, &resp)

			assert.Equal(t, tc.deleteStatus != 0, deleted)
			if tc.err == "" {
				require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
				return
			}
			require.True(t, resp.Diagnostics.HasError())
			assert.Equal(t, tc.err, resp.Diagnostics.Errors()[0].Summary())
		})
	}
}

func TestTagsKeepConfiguredForm(t *testing.T) {
	empty := types.SetValueMust(types.StringType, []attr.Value{})
	for name, tc := range map[string]struct {
		configured types.Set
		remote     []string
		want       types.Set
	}{
		"omitted":         {types.SetNull(types.StringType), nil, types.SetNull(types.StringType)},
		"empty":           {empty, nil, empty},
		"removed outside": {types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")}), nil, empty},
		"remote set":      {types.SetNull(types.StringType), []string{"a"}, types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
	} {
		got, diags := tagsValue(tc.configured, tc.remote)
		require.False(t, diags.HasError())
		assert.Equal(t, tc.want, got, name)
	}
}

func TestReadImage(t *testing.T) {
	// The image changed outside of Terraform in every attribute that the API reports.
	changed := apiImage(map[string]any{
		"name": "renamed outside", "visibility": "shared", "protected": true, "os_hidden": true,
		"min_disk": 2, "min_ram": 512, "owner": memberID, "size": 7, "checksum": "md5",
		"container_format": "ova", "disk_format": "qcow2", "status": "deactivated", "tags": []string{"x"},
		"created_at": "2026-10-10T10:00:00Z", "updated_at": "2026-10-11T10:00:00Z",
	})
	prior := planned()
	prior.ID = types.StringValue(imageID)
	refreshed := prior
	refreshed.Name, refreshed.Visibility = types.StringValue("renamed outside"), types.StringValue("shared")
	refreshed.Protected, refreshed.Hidden = types.BoolValue(true), types.BoolValue(true)
	refreshed.MinDiskGB, refreshed.MinRAMMB = types.Int64Value(2), types.Int64Value(512)
	refreshed.Owner, refreshed.SizeBytes, refreshed.Checksum = types.StringValue(memberID), types.Int64Value(7), types.StringValue("md5")
	refreshed.ContainerFormat, refreshed.DiskFormat = types.StringValue("ova"), types.StringValue("qcow2")
	refreshed.Status = types.StringValue("deactivated")
	refreshed.Tags = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("x")})
	// The properties stay unmanaged: the configuration has none.
	refreshed.ServiceProperties = properties(map[string]string{"os_distro": "ubuntu"})
	refreshed.CreatedAt, refreshed.UpdatedAt = types.StringValue("2026-10-10T10:00:00Z"), types.StringValue("2026-10-11T10:00:00Z")
	for name, tc := range map[string]struct {
		status int
		image  map[string]any
		gone   bool
		err    bool
	}{
		"changed":        {http.StatusOK, changed, false, false},
		"deleted":        {http.StatusOK, apiImage(map[string]any{"status": "deleted"}), true, false},
		"pending_delete": {http.StatusOK, apiImage(map[string]any{"status": "pending_delete"}), true, false},
		"not found":      {http.StatusNotFound, map[string]any{"message": "not found"}, true, false},
		// Another error keeps the image in the state, so no duplicate is created.
		"server error": {http.StatusInternalServerError, map[string]any{"message": "error"}, false, true},
	} {
		t.Run(name, func(t *testing.T) {
			r := testImageResource(t, func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, tc.status, tc.image)
			})
			resp := resource.ReadResponse{State: imageState(t, r, prior)}
			r.Read(context.Background(), resource.ReadRequest{State: imageState(t, r, prior)}, &resp)
			assert.Equal(t, tc.gone, resp.State.Raw.IsNull())
			assert.Equal(t, tc.err, resp.Diagnostics.HasError())
			if !tc.gone && !tc.err {
				assert.Equal(t, refreshed, readImage(t, resp.State))
			}
		})
	}
}

func TestImportImage(t *testing.T) {
	configured := testImageResource(t, nil)
	for name, tc := range map[string]struct {
		r   *imageResource
		id  string
		err string
	}{
		"image ID":        {configured, imageID, ""},
		"uppercase ID":    {configured, "AAAAAAAA-1111-4111-8111-111111111111", "Invalid image ID"},
		"not a UUID":      {configured, "ubuntu", "Invalid image ID"},
		"no import scope": {&imageResource{}, imageID, "Missing provider configuration"},
	} {
		t.Run(name, func(t *testing.T) {
			var s resource.SchemaResponse
			tc.r.Schema(context.Background(), resource.SchemaRequest{}, &s)
			resp := resource.ImportStateResponse{State: tfsdk.State{
				Schema: s.Schema, Raw: tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil),
			}}
			tc.r.ImportState(context.Background(), resource.ImportStateRequest{ID: tc.id}, &resp)
			if tc.err != "" {
				require.True(t, resp.Diagnostics.HasError())
				assert.Equal(t, tc.err, resp.Diagnostics.Errors()[0].Summary())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			imported := readImage(t, resp.State)
			assert.Equal(t, imageID, imported.ID.ValueString())
			assert.Equal(t, projectID, imported.ProjectID.ValueString())
		})
	}
}
