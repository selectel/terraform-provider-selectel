package blockstorage

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dataConfig(t *testing.T, source datasource.DataSource, model any) tfsdk.Config {
	t.Helper()
	var schema datasource.SchemaResponse
	source.Schema(context.Background(), datasource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	d := state.Set(context.Background(), model)
	require.False(t, d.HasError(), "%v", d)

	return tfsdk.Config{Schema: schema.Schema, Raw: state.Raw}
}

func TestVolumeLookup(t *testing.T) {
	for _, scenario := range []string{"id", "filter", "empty value", "missing key", "no match", "ambiguous", "get error", "list error", "pagination"} {
		t.Run(scenario, func(t *testing.T) {
			item := apiVolume()
			item.Bootable = "true"
			item.Metadata = map[string]string{"team": "infra", "empty": "", createTokenKey: "token"}
			client := testClient(t, volumeLookupHandler(t, scenario, item))
			source := NewVolumeDataSource().(*volumeDataSource)
			source.config = testConfig{client: client}
			lookup := volumeDataModel{
				ProjectID:   types.StringValue("project"),
				Region:      types.StringValue("ru-1"),
				Metadata:    types.MapNull(types.StringType),
				Attachments: types.SetNull(types.ObjectType{AttrTypes: attachmentTypes}),
			}
			if scenario == "id" || scenario == "get error" {
				lookup.VolumeID = types.StringValue(volumeID)
			} else {
				lookup.Name = types.StringValue("disk")
				lookup.Status = types.StringValue("available")
				lookup.Metadata = metadata(map[string]string{"team": "infra"})
			}
			if scenario == "empty value" {
				lookup.Metadata = metadata(map[string]string{"empty": ""})
			}
			if scenario == "missing key" {
				lookup.Metadata = metadata(map[string]string{"missing": ""})
			}
			config := dataConfig(t, source, lookup)
			response := datasource.ReadResponse{State: tfsdk.State{Schema: config.Schema}}
			source.Read(context.Background(), datasource.ReadRequest{Config: config}, &response)
			failure := scenario == "no match" || scenario == "ambiguous" ||
				scenario == "get error" ||
				scenario == "list error" ||
				scenario == "missing key"
			require.Equal(t, failure, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			if failure {
				assert.True(t, response.State.Raw.IsNull())
				if scenario == "list error" || scenario == "get error" {
					require.Len(t, response.Diagnostics.Errors(), 1)
					if scenario == "list error" {
						assert.Contains(t, response.Diagnostics.Errors()[0].Detail(), "incomplete_list")
					} else {
						assert.Contains(t, response.Diagnostics.Errors()[0].Detail(), "status")
					}
				}

				return
			}
			var state volumeDataModel
			require.False(t, response.State.Get(context.Background(), &state).HasError())
			assert.Equal(t, volumeID, state.ID.ValueString())
			assert.Equal(t, "disk", state.Name.ValueString())
			assert.Equal(t, "available", state.Status.ValueString())
			assert.Equal(t, int64(10), state.Size.ValueInt64())
			assert.Equal(t, "description", state.Description.ValueString())
			assert.True(t, state.Bootable.ValueBool())
			assert.Equal(t, "fast.ru-1a", state.VolumeType.ValueString())
			assert.Equal(t, types.StringValue("ru-1a"), state.AvailabilityZone)
			assert.Equal(t, types.StringValue(""), state.SourceVolumeID)
			assert.Equal(t, types.StringValue(""), state.SnapshotID)
			assert.Equal(t, types.SetValueMust(types.ObjectType{AttrTypes: attachmentTypes}, []attr.Value{}), state.Attachments)
			assert.NotContains(t, state.Metadata.Elements(), createTokenKey)
			assert.Len(t, state.Metadata.Elements(), 2)
		})
	}
}

func TestSnapshotLookup(t *testing.T) {
	for _, scenario := range []string{"single", "unfiltered", "newest", "tie", "ambiguous", "empty", "error", "pagination", "case-insensitive API"} {
		t.Run(scenario, func(t *testing.T) {
			testSnapshotLookup(t, scenario)
		})
	}
}

func TestVolumeTypeLookup(t *testing.T) {
	for _, scenario := range []string{"id", "default", "name", "ambiguous", "absent", "get error", "list error", "qos error", "qos disabled", "partial qos", "other type"} {
		t.Run(scenario, func(t *testing.T) {
			extraSpecs := map[string]string{"key": "value"}
			item := map[string]any{
				"id":          typeID,
				"name":        "fast.ru-1a",
				"description": "description",
				"is_public":   true,
				"extra_specs": extraSpecs,
			}
			client := testClient(t, volumeTypeLookupHandler(t, scenario, item))
			source := NewVolumeTypeDataSource().(*volumeTypeDataSource)
			source.config = testConfig{client: client}
			lookup := volumeTypeDataModel{
				ProjectID:  types.StringValue("project"),
				Region:     types.StringValue("ru-1"),
				ExtraSpecs: types.MapNull(types.StringType),
			}
			if scenario == "name" || scenario == "ambiguous" || scenario == "absent" || scenario == "list error" {
				lookup.Name = types.StringValue("fast.ru-1a")
			} else {
				lookup.TypeID = types.StringValue(typeID)
			}
			if scenario == "default" {
				lookup.TypeID = types.StringValue("default")
				delete(item, "extra_specs")
				extraSpecs = map[string]string{}
			}
			config := dataConfig(t, source, lookup)
			response := datasource.ReadResponse{State: tfsdk.State{Schema: config.Schema}}
			source.Read(context.Background(), datasource.ReadRequest{Config: config}, &response)
			failure := scenario == "ambiguous" || scenario == "absent" || scenario == "get error" ||
				scenario == "list error" ||
				scenario == "qos error"
			require.Equal(t, failure, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			if failure {
				assert.True(t, response.State.Raw.IsNull())
				if scenario == "list error" || scenario == "get error" || scenario == "qos error" {
					require.Len(t, response.Diagnostics.Errors(), 1)
					if scenario == "list error" {
						assert.Contains(t, response.Diagnostics.Errors()[0].Detail(), "incomplete_list")
					} else {
						assert.Contains(t, response.Diagnostics.Errors()[0].Detail(), "status")
					}
				}

				return
			}
			var state volumeTypeDataModel
			require.False(t, response.State.Get(context.Background(), &state).HasError())
			assert.Equal(t, typeID, state.ID.ValueString())
			assert.Equal(t, "fast.ru-1a", state.Name.ValueString())
			assert.Equal(t, "description", state.Description.ValueString())
			assert.True(t, state.IsPublic.ValueBool())
			assert.Equal(t, metadata(extraSpecs), state.ExtraSpecs)
			assert.Equal(
				t,
				scenario != "qos disabled" && scenario != "partial qos" && scenario != "other type",
				state.CustomIOPS.ValueBool(),
			)
		})
	}
}

func volumeLookupHandler(t *testing.T, scenario string, item volume.View) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/volumes/"+volumeID {
			if scenario == "get error" {
				w.WriteHeader(403)
				return
			}
			writeJSON(t, w, 200, map[string]any{"volume": item})

			return
		}
		assert.Equal(t, "/volumes/detail", req.URL.Path)
		if scenario == "list error" {
			w.WriteHeader(403)
			return
		}
		all := make([]volume.View, 1, 4)
		all[0] = item
		switch scenario {
		case "filter":
			otherName, otherStatus := item, item
			otherName.Name = "other"
			otherStatus.Status = "in-use"
			all[0] = otherName
			all = append(all, otherStatus, item)
		case "no match":
			all[0].Name = "other"
		case "ambiguous":
			all = append(all, item)
		}
		if scenario == "pagination" && req.URL.Query().Get("marker") == "" {
			writeJSON(
				t,
				w,
				200,
				map[string]any{
					"volumes": []volume.View{},
					"volumes_links": []map[string]string{
						{"rel": "next", "href": "http://" + req.Host + "/volumes/detail?marker=page2"},
					},
				},
			)

			return
		}
		writeJSON(t, w, 200, map[string]any{"volumes": all})
	}
}

func volumeTypeLookupHandler(t *testing.T, scenario string, item map[string]any) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/types/" + typeID, "/types/default":
			if scenario == "get error" {
				w.WriteHeader(404)
				return
			}
			writeJSON(t, w, 200, map[string]any{"volume_type": item})
		case "/types":
			if scenario == "list error" {
				w.WriteHeader(500)
				return
			}
			all := []any{item}
			if scenario == "ambiguous" {
				all = append(all, item)
			}
			if scenario == "absent" {
				all = []any{}
			}
			writeJSON(t, w, 200, map[string]any{"volume_types": all})
		case "/qos-specs/qos_limits":
			if scenario == "qos error" {
				w.WriteHeader(403)
				return
			}
			id := typeID
			if scenario == "other type" {
				id = volumeID
			}
			writeJSON(
				t,
				w,
				200,
				map[string]any{
					"fast": map[string]any{
						"volume_type_id":     id,
						"allow_user_qos":     scenario != "qos disabled",
						"full_qos_disk_type": scenario != "partial qos",
					},
				},
			)
		default:
			t.Errorf("unexpected request %s", req.URL.Path)
			w.WriteHeader(400)
		}
	}
}

func testSnapshotLookup(t *testing.T, scenario string) {
	t.Helper()
	snapshot := func(id, created string) map[string]any {
		return map[string]any{
			"id":          id,
			"created_at":  created,
			"name":        "snapshot",
			"status":      "available",
			"volume_id":   volumeID,
			"description": "description",
			"size":        10,
			"metadata":    map[string]string{"key": "value"},
		}
	}
	client := testClient(t, func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/snapshots", req.URL.Path)
		if scenario == "error" {
			w.WriteHeader(500)
			return
		}
		if req.URL.Query().Get("marker") == "" && scenario != "unfiltered" {
			assert.Equal(t, "snapshot", req.URL.Query().Get("name"))
			assert.Equal(t, volumeID, req.URL.Query().Get("volume_id"))
			assert.Equal(t, "available", req.URL.Query().Get("status"))
		}
		all := []any{snapshot("b", "2026-10-01T10:00:00")}
		switch scenario {
		case "newest":
			all = append(all, snapshot("c", "2026-10-02T10:00:00"), snapshot("a", "2026-09-30T10:00:00"))
		case "tie":
			all = append(all, snapshot("a", "2026-10-01T10:00:00"), snapshot("z", "2026-10-01T10:00:00"))
		case "ambiguous":
			all = append(all, snapshot("c", "2026-10-02T10:00:00"))
		case "empty":
			all = all[:0]
		case "case-insensitive API":
			// The API filter matches names case-insensitively.
			other := snapshot("c", "2026-10-02T10:00:00")
			other["name"] = "Snapshot"
			all = append(all, other)
		}
		if scenario == "pagination" && req.URL.Query().Get("marker") == "" {
			writeJSON(
				t,
				w,
				200,
				map[string]any{
					"snapshots": []any{},
					"snapshots_links": []map[string]string{
						{"rel": "next", "href": "http://" + req.Host + "/snapshots?marker=page2"},
					},
				},
			)

			return
		}
		writeJSON(t, w, 200, map[string]any{"snapshots": all})
	})
	source := NewSnapshotDataSource().(*snapshotDataSource)
	source.config = testConfig{client: client}
	lookup := snapshotDataModel{
		ProjectID:  types.StringValue("project"),
		Region:     types.StringValue("ru-1"),
		Name:       types.StringValue("snapshot"),
		Status:     types.StringValue("available"),
		VolumeID:   types.StringValue(volumeID),
		MostRecent: types.BoolValue(scenario == "newest" || scenario == "tie"),
		Metadata:   types.MapNull(types.StringType),
	}
	if scenario == "unfiltered" {
		lookup.Name = types.StringNull()
		lookup.Status = types.StringNull()
		lookup.VolumeID = types.StringNull()
	}
	config := dataConfig(t, source, lookup)
	response := datasource.ReadResponse{State: tfsdk.State{Schema: config.Schema}}
	source.Read(context.Background(), datasource.ReadRequest{Config: config}, &response)
	failure := scenario == "ambiguous" || scenario == "empty" || scenario == "error"
	require.Equal(t, failure, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	if failure {
		return
	}
	var state snapshotDataModel
	require.False(t, response.State.Get(context.Background(), &state).HasError())
	id := "b"
	if scenario == "newest" {
		id = "c"
	}
	if scenario == "tie" {
		id = "a"
	}
	assert.Equal(t, id, state.ID.ValueString())
	assert.Equal(t, "snapshot", state.Name.ValueString())
	assert.Equal(t, "available", state.Status.ValueString())
	assert.Equal(t, volumeID, state.VolumeID.ValueString())
	assert.Equal(t, "description", state.Description.ValueString())
	assert.Equal(t, int64(10), state.Size.ValueInt64())
	assert.Equal(t, metadata(map[string]string{"key": "value"}), state.Metadata)
}
