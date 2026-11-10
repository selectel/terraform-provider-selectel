package blockstorage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

type volumeDataSource struct{ dataConnection }

func NewVolumeDataSource() datasource.DataSource { return &volumeDataSource{} }

func (d *volumeDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_compute_volume_v3"
}

type volumeDataModel struct {
	ID               types.String `tfsdk:"id"`
	ProjectID        types.String `tfsdk:"project_id"`
	Region           types.String `tfsdk:"region"`
	VolumeID         types.String `tfsdk:"volume_id"`
	Name             types.String `tfsdk:"name"`
	Status           types.String `tfsdk:"status"`
	Metadata         types.Map    `tfsdk:"metadata"`
	Description      types.String `tfsdk:"description"`
	Size             types.Int64  `tfsdk:"size"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	VolumeType       types.String `tfsdk:"volume_type"`
	SnapshotID       types.String `tfsdk:"snapshot_id"`
	SourceVolumeID   types.String `tfsdk:"source_volume_id"`
	Bootable         types.Bool   `tfsdk:"bootable"`
	Attachments      types.Set    `tfsdk:"attachments"`
}

func (d *volumeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	docs := schemadoc.DataSource{Name: "network volume", Location: volumeLocation}
	byID := schemadoc.ConflictsWith("volume_id")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a network volume in Cloud Servers. The data source finds the volume by `volume_id` or by any combination of `name`, `status`, and `metadata`. The search must return exactly one network volume. If `volume_id` and search criteria are not specified, there must be exactly one network volume in the specified project and pool. For more information about network volumes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/volumes/about-network-volumes/).",
		Attributes: map[string]schema.Attribute{
			"id":         docs.ID(),
			"project_id": docs.ProjectID(),
			"region":     docs.Region(),
			"volume_id": docs.String(schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{schemadoc.UUID(), schemadoc.ConflictsWith("name", "status", "metadata")},
			}, schemadoc.Field{
				Definition: "Unique identifier of the network volume.",
				Source:     schemadoc.FromOpenStackCLI("volume ID", "openstack volume list", "ID"),
			}),
			"name": docs.String(filterString(byID), schemadoc.Field{
				Definition: "Exact network volume name to search for.",
				Behavior:   "If omitted, the data source returns the name of the selected volume.",
				Source:     schemadoc.FromOpenStackCLI("volume name", "openstack volume list", "Name"),
			}),
			"status": docs.String(filterString(byID), schemadoc.Field{
				Definition: "Exact API status to search for.",
				Behavior:   "If omitted, the data source returns the status of the selected volume.",
				Source:     schemadoc.FromOpenStackCLI("volume status", "openstack volume list", "Status"),
			}),
			"metadata": docs.Map(schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators:  []validator.Map{byID},
			}, schemadoc.Field{
				Definition: "Metadata subset to match.",
				Behavior:   "The data source returns all metadata of the selected volume except the provider creation token.",
			}),
			"description":       schema.StringAttribute{Computed: true, Description: "Volume description."},
			"size":              schema.Int64Attribute{Computed: true, Description: "Size in GB."},
			"availability_zone": schema.StringAttribute{Computed: true, Description: "Pool segment."},
			"volume_type":       schema.StringAttribute{Computed: true, Description: "Volume type name returned by the API."},
			"snapshot_id":       schema.StringAttribute{Computed: true, Description: "Unique identifier of the source snapshot. The value is empty if the volume is not created from a snapshot."},
			"source_volume_id":  schema.StringAttribute{Computed: true, Description: "Unique identifier of the source network volume. The value is empty if the volume is not cloned from another volume."},
			"bootable":          schema.BoolAttribute{Computed: true, Description: "Whether the volume is bootable."},
			"attachments": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Volume attachments.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, Description: "Unique identifier of the network volume."},
						"instance_id": schema.StringAttribute{Computed: true, Description: "Unique identifier of the attached cloud server."},
						"device":      schema.StringAttribute{Computed: true, Description: "Device name on the cloud server."},
					},
				},
			},
		},
	}
}

func (d *volumeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeDataModel
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
	var selected *volume.View
	if !config.VolumeID.IsNull() {
		selected, _, err = volume.Get(ctx, client, config.VolumeID.ValueString())
	} else {
		var all []volume.View
		all, err = volume.ListDetail(ctx, client, volume.ListOpts{})
		filters := map[string]string{}
		if !config.Metadata.IsNull() {
			resp.Diagnostics.Append(config.Metadata.ElementsAs(ctx, &filters, false)...)
		}
		if resp.Diagnostics.HasError() {
			return
		}
		matches := 0
		for i := range all {
			candidate := &all[i]
			if !matchesFilter(config.Name, candidate.Name) || !matchesFilter(config.Status, candidate.Status) {
				continue
			}
			match := true
			for key, value := range filters {
				v, exists := candidate.Metadata[key]
				if !exists || v != value {
					match = false
					break
				}
			}
			if match {
				selected = candidate
				matches++
			}
		}
		if err == nil && matches != 1 {
			err = fmt.Errorf("expected exactly one matching volume, found %d; refine filters or use volume_id", matches)
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Find volume", err.Error())
		return
	}
	bootable := false
	if selected.Bootable != "" {
		bootable, err = strconv.ParseBool(selected.Bootable)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid volume response",
			fmt.Sprintf("invalid bootable value %q", selected.Bootable),
		)

		return
	}
	config.ID = types.StringValue(selected.ID)
	config.Name = types.StringValue(selected.Name)
	config.Status = types.StringValue(selected.Status)
	config.Description = types.StringValue(selected.Description)
	config.Size = types.Int64Value(int64(selected.Size))
	config.AvailabilityZone = types.StringValue(selected.AvailabilityZone)
	config.VolumeType = types.StringValue(selected.VolumeType)
	config.SnapshotID = types.StringValue(selected.SnapshotID)
	config.SourceVolumeID = types.StringValue(selected.SourceVolID)
	config.Bootable = types.BoolValue(bootable)
	var diags diag.Diagnostics
	config.Metadata, diags = types.MapValueFrom(ctx, types.StringType, serviceMetadata(selected))
	resp.Diagnostics.Append(diags...)
	config.Attachments, diags = attachmentsValue(selected.Attachments)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
