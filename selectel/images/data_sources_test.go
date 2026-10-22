package images

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func emptyFilter() imageFilter {
	return imageFilter{
		Name: types.StringNull(), NameRegex: types.StringNull(), Visibility: types.StringNull(),
		MemberStatus: types.StringNull(), Owner: types.StringNull(),
		SizeMin: types.Int64Null(), SizeMax: types.Int64Null(),
		Tag: types.StringNull(), Tags: types.SetNull(types.StringType), Properties: types.MapNull(types.StringType),
		Hidden: types.BoolNull(), ContainerFormat: types.StringNull(), DiskFormat: types.StringNull(),
	}
}

// listing serves the images as one page and records the query of the list request.
func listing(t *testing.T, query *url.Values, images ...map[string]any) dataConnection {
	t.Helper()
	if images == nil {
		images = []map[string]any{}
	}

	return dataConnection{testConnection(t, func(w http.ResponseWriter, req *http.Request) {
		*query = req.URL.Query()
		writeJSON(t, w, http.StatusOK, map[string]any{"images": images})
	})}
}

func readDataSource(t *testing.T, d datasource.DataSource, config any) datasource.ReadResponse {
	t.Helper()
	ctx := context.Background()
	var s datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &s)
	state := tfsdk.State{Schema: s.Schema}
	require.False(t, state.Set(ctx, config).HasError())
	resp := datasource.ReadResponse{State: state}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config(state)}, &resp)

	return resp
}

func TestFindImagesMatchesExactly(t *testing.T) {
	// The API compares names and tags case-insensitively, and name_regex,
	// owner and formats are checked by the provider.
	images := []map[string]any{
		apiImage(map[string]any{"id": "1", "name": "Ubuntu", "tags": []string{"a", "b"}}),
		apiImage(map[string]any{"id": "2", "name": "ubuntu", "tags": []string{"a", "b"}}),
		apiImage(map[string]any{"id": "3", "name": "ubuntu", "tags": []string{"A", "b"}}),
		apiImage(map[string]any{"id": "4", "name": "ubuntu", "tags": []string{"a", "b"}, "os_distro": "debian"}),
		apiImage(map[string]any{"id": "5", "name": "ubuntu", "tags": []string{"a", "b"}, "owner": memberID}),
		apiImage(map[string]any{"id": "6", "name": "ubuntu", "tags": []string{"a", "b"}, "disk_format": "qcow2"}),
		apiImage(map[string]any{"id": "7", "name": "ubuntu", "tags": []string{"a", "b"}, "container_format": "ova"}),
	}
	for name, tc := range map[string]struct {
		byName func(*imageIDsDataModel)
		sort   types.String
		hidden types.Bool
		// query is the sort and the os_hidden filter of the list request.
		query url.Values
	}{
		"name": {
			func(m *imageIDsDataModel) { m.Name = types.StringValue("ubuntu") },
			types.StringNull(), types.BoolNull(),
			url.Values{"sort": {"name:asc"}},
		},
		"name_regex": {
			func(m *imageIDsDataModel) { m.NameRegex = types.StringValue("^ubuntu$") },
			types.StringValue("created_at:desc"), types.BoolValue(false),
			url.Values{"sort": {"created_at:desc"}, "os_hidden": {"false"}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var query url.Values
			d := &imageIDsDataSource{listing(t, &query, images...)}
			config := imageIDsDataModel{
				ID: types.StringNull(), ProjectID: types.StringValue(projectID), Region: types.StringValue("ru-6"),
				imageFilter: emptyFilter(), Sort: types.StringNull(), IDs: types.ListNull(types.StringType),
			}
			tc.byName(&config)
			config.Sort, config.Hidden = tc.sort, tc.hidden
			config.Owner = types.StringValue(projectID)
			config.ContainerFormat, config.DiskFormat = types.StringValue("bare"), types.StringValue("raw")
			config.Tag = types.StringValue("a")
			config.Tags = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("b")})
			config.Properties = properties(map[string]string{"os_distro": "ubuntu"})

			resp := readDataSource(t, d, &config)
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var found imageIDsDataModel
			require.False(t, resp.State.Get(context.Background(), &found).HasError())
			assert.Equal(t, types.ListValueMust(types.StringType, []attr.Value{types.StringValue("2")}), found.IDs)
			digest := sha256.Sum256([]byte("2"))
			assert.Equal(t, hex.EncodeToString(digest[:]), found.ID.ValueString())
			assert.ElementsMatch(t, []string{"b", "a"}, query["tag"])
			// hidden is sent only when set.
			for key, want := range map[string][]string{"sort": tc.query["sort"], "os_hidden": tc.query["os_hidden"]} {
				assert.Equal(t, want, query[key], key)
			}
		})
	}
}

func TestImageDataSourceSelectsOne(t *testing.T) {
	older := apiImage(map[string]any{
		"id": "a", "created_at": "2026-10-01T00:00:00Z", "updated_at": "2026-10-03T00:00:00Z",
		"name": "ubuntu", "visibility": "shared", "owner": memberID, "os_hidden": true, "protected": true,
		"container_format": "ova", "disk_format": "qcow2", "tags": []string{"t"},
		"min_disk": 2, "min_ram": 512, "size": 7, "checksum": "md5",
	})
	newer := apiImage(map[string]any{"id": "c", "created_at": "2026-10-02T00:00:00Z"})
	sameTime := apiImage(map[string]any{"id": "b", "created_at": "2026-10-02T00:00:00Z"})
	sameTimeLater := apiImage(map[string]any{"id": "d", "created_at": "2026-10-02T00:00:00Z"})
	for name, tc := range map[string]struct {
		images     []map[string]any
		mostRecent bool
		selected   string
		err        string
	}{
		"single":      {[]map[string]any{older}, false, "a", ""},
		"none":        {nil, false, "", "Image not found"},
		"several":     {[]map[string]any{older, newer}, false, "", "Ambiguous image lookup"},
		"most recent": {[]map[string]any{older, newer}, true, "c", ""},
		// Images created at the same time are ordered by ID.
		"most recent, same time": {[]map[string]any{newer, sameTime, sameTimeLater, older}, true, "b", ""},
	} {
		t.Run(name, func(t *testing.T) {
			var query url.Values
			d := &imageDataSource{listing(t, &query, tc.images...)}
			config := imageDataModel{
				ID: types.StringNull(), ProjectID: types.StringValue(projectID), Region: types.StringValue("ru-6"),
				imageFilter: emptyFilter(), MostRecent: types.BoolValue(tc.mostRecent),
				Protected: types.BoolNull(), MinDiskGB: types.Int64Null(), MinRAMMB: types.Int64Null(),
				SizeBytes: types.Int64Null(), Checksum: types.StringNull(),
				CreatedAt: types.StringNull(), UpdatedAt: types.StringNull(),
			}
			resp := readDataSource(t, d, &config)
			if tc.err != "" {
				require.True(t, resp.Diagnostics.HasError())
				assert.Equal(t, tc.err, resp.Diagnostics.Errors()[0].Summary())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var selected imageDataModel
			require.False(t, resp.State.Get(context.Background(), &selected).HasError())
			assert.Equal(t, tc.selected, selected.ID.ValueString())
			if tc.selected != "a" {
				return
			}
			// The filters that describe the image take its values.
			want := config
			want.ID, want.Name, want.Visibility = types.StringValue("a"), types.StringValue("ubuntu"), types.StringValue("shared")
			want.Owner, want.Hidden, want.Protected = types.StringValue(memberID), types.BoolValue(true), types.BoolValue(true)
			want.ContainerFormat, want.DiskFormat = types.StringValue("ova"), types.StringValue("qcow2")
			want.Tags = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("t")})
			want.Properties = properties(map[string]string{"os_distro": "ubuntu"})
			want.MinDiskGB, want.MinRAMMB, want.SizeBytes = types.Int64Value(2), types.Int64Value(512), types.Int64Value(7)
			want.Checksum = types.StringValue("md5")
			want.CreatedAt, want.UpdatedAt = types.StringValue("2026-10-01T00:00:00Z"), types.StringValue("2026-10-03T00:00:00Z")
			assert.Equal(t, want, selected)
		})
	}
}
