package blockstorage

import (
	"context"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/blockstorage-go/pkg/v1/snapshot"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

type snapshotDataSource struct{ dataConnection }

func NewSnapshotDataSource() datasource.DataSource { return &snapshotDataSource{} }

func (d *snapshotDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_compute_snapshot_v3"
}

type snapshotDataModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Region      types.String `tfsdk:"region"`
	Name        types.String `tfsdk:"name"`
	Status      types.String `tfsdk:"status"`
	VolumeID    types.String `tfsdk:"volume_id"`
	MostRecent  types.Bool   `tfsdk:"most_recent"`
	Description types.String `tfsdk:"description"`
	Size        types.Int64  `tfsdk:"size"`
	Metadata    types.Map    `tfsdk:"metadata"`
}

func (d *snapshotDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	docs := schemadoc.DataSource{Name: "snapshot", Location: volumeLocation}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a network volume snapshot in Cloud Servers. The data source filters snapshots by name, status, or source volume. Multiple matches cause an error unless `most_recent` is `true`. For more information about snapshots, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/volumes/snapshots/).",
		Attributes: map[string]schema.Attribute{
			"id":         docs.ID(),
			"project_id": docs.ProjectID(),
			"region":     docs.Region(),
			"name": docs.String(filterString(), schemadoc.Field{
				Definition: "Exact snapshot name to search for.",
				Behavior:   "If omitted, the data source returns the name of the selected snapshot.",
				Source:     schemadoc.FromOpenStackCLI("snapshot name", "openstack volume snapshot list", "Name"),
			}),
			"status": docs.String(filterString(), schemadoc.Field{
				Definition: "Exact API status to search for.",
				Behavior:   "Use a lowercase value, for example, `available`. If omitted, the data source returns the status of the selected snapshot.",
				Source:     schemadoc.FromOpenStackCLI("snapshot status", "openstack volume snapshot list", "Status"),
				Concept: schemadoc.Link{
					Title: "snapshot statuses",
					URL:   "https://docs.selectel.ru/en/cloud-servers/volumes/snapshots/#snapshot-statuses",
				},
			}),
			"volume_id": docs.String(schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{schemadoc.UUID()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the source network volume.",
				Behavior:   "The data source filters snapshots by this source volume. If omitted, it returns the source volume ID of the selected snapshot.",
				Source:     schemadoc.FromDataSource("selectel_compute_volume_v3", "source volume ID"),
			}),
			"most_recent": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Specifies if the most recent matching snapshot is selected. If omitted or `false`, multiple matches cause an error. Equal timestamps are resolved by the lexicographically smallest UUID.",
			},
			"description": schema.StringAttribute{Computed: true, Description: "Snapshot description."},
			"size":        schema.Int64Attribute{Computed: true, Description: "Snapshot size in GB."},
			"metadata":    schema.MapAttribute{Computed: true, ElementType: types.StringType, Description: "Snapshot metadata returned by the API."},
		},
	}
}

func (d *snapshotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config snapshotDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	client, err := d.client(ctx, config.ProjectID.ValueString(), config.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to Block Storage", err.Error())
		return
	}
	all, err := snapshot.List(
		ctx,
		client,
		snapshot.ListOpts{
			Name:     config.Name.ValueString(),
			Status:   config.Status.ValueString(),
			VolumeID: config.VolumeID.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Find snapshot", err.Error())
		return
	}
	// The API compares these filters case-insensitively, so the request only
	// narrows the list and the configured values must still match exactly.
	all = slices.DeleteFunc(all, func(candidate snapshot.View) bool {
		return !matchesFilter(config.Name, candidate.Name) ||
			!matchesFilter(config.Status, candidate.Status) ||
			!matchesFilter(config.VolumeID, candidate.VolumeID)
	})
	if len(all) == 0 {
		resp.Diagnostics.AddError("Snapshot not found", "No snapshots matched the configured filters.")
		return
	}
	if len(all) > 1 && !config.MostRecent.ValueBool() {
		resp.Diagnostics.AddError("Ambiguous snapshot lookup", "Refine filters or set most_recent = true.")
		return
	}
	selected := all[0]
	for _, candidate := range all[1:] {
		if candidate.CreatedAt.After(selected.CreatedAt) ||
			(candidate.CreatedAt.Equal(selected.CreatedAt) && candidate.ID < selected.ID) {
			selected = candidate
		}
	}
	config.ID = types.StringValue(selected.ID)
	config.Name = types.StringValue(selected.Name)
	config.Status = types.StringValue(selected.Status)
	config.VolumeID = types.StringValue(selected.VolumeID)
	config.Description = types.StringValue(selected.Description)
	config.Size = types.Int64Value(int64(selected.Size))
	var diags diag.Diagnostics
	config.Metadata, diags = types.MapValueFrom(ctx, types.StringType, selected.Metadata)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
