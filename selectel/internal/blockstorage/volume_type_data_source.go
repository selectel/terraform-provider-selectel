package blockstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/blockstorage-go/pkg/v1/volumetype"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

type volumeTypeDataSource struct{ dataConnection }

const volumeTypeNameDescription = "The name format is `<volume_type>.<availability_zone>`, for example, `fast2.ru-6a` for SSD Fast v2 in the `ru-6a` pool segment."

var volumeTypesConcept = schemadoc.Link{
	Title: "the list of network volume types",
	URL:   "https://docs.selectel.ru/en/cloud-servers/volumes/about-network-volumes/#network-volume-types-list",
}

func NewVolumeTypeDataSource() datasource.DataSource { return &volumeTypeDataSource{} }

func (d *volumeTypeDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_compute_volume_type_v3"
}

type volumeTypeDataModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Region      types.String `tfsdk:"region"`
	TypeID      types.String `tfsdk:"volume_type_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsPublic    types.Bool   `tfsdk:"is_public"`
	ExtraSpecs  types.Map    `tfsdk:"extra_specs"`
	CustomIOPS  types.Bool   `tfsdk:"supports_custom_iops"`
}

func (d *volumeTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	docs := schemadoc.DataSource{Name: "volume type", Location: volumeLocation}
	selectors := schemadoc.ExactlyOneOf("volume_type_id", "name")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a network volume type in Cloud Servers. The data source finds exactly one type by UUID, `default`, or exact name. For more information about volume types, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/volumes/about-network-volumes/#network-volume-types).",
		Attributes: map[string]schema.Attribute{
			"id":         docs.ID(),
			"project_id": docs.ProjectID(),
			"region":     docs.Region(),
			"volume_type_id": docs.String(schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{volumeTypeIDValidator(), selectors},
			}, schemadoc.Field{
				Definition: "Unique identifier of the volume type.",
				Source:     schemadoc.FromOpenStackCLI("volume type ID", "openstack volume type list", "ID"),
			}),
			"name": docs.String(filterString(selectors), schemadoc.Field{
				Definition: "Exact volume type name.",
				Behavior:   volumeTypeNameDescription + " ID lookups return the API name here.",
				Source:     schemadoc.FromOpenStackCLI("volume type name", "openstack volume type list", "Name"),
				Concept:    volumeTypesConcept,
			}),
			"description": schema.StringAttribute{Computed: true, Description: "Volume type description."},
			"is_public":   schema.BoolAttribute{Computed: true, Description: "Whether the type is public."},
			"extra_specs": schema.MapAttribute{Computed: true, ElementType: types.StringType, Description: "Type specifications visible to the current role."},
			"supports_custom_iops": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the type allows custom `total_iops_sec` metadata.",
			},
		},
	}
}

func (d *volumeTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeTypeDataModel
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
	var selected *volumetype.View
	if !config.TypeID.IsNull() {
		selected, _, err = volumetype.Get(ctx, client, config.TypeID.ValueString())
	} else {
		var all []volumetype.View
		all, err = volumetype.List(ctx, client, volumetype.ListOpts{})
		matches := 0
		for i := range all {
			if all[i].Name == config.Name.ValueString() {
				selected = &all[i]
				matches++
			}
		}
		if err == nil && matches != 1 {
			err = fmt.Errorf("expected exactly one volume type named %q, found %d", config.Name.ValueString(), matches)
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Find volume type", err.Error())
		return
	}
	limits, _, err := volumetype.ListQoSLimits(ctx, client)
	if err != nil {
		resp.Diagnostics.AddError("Read volume type QoS", err.Error())
		return
	}
	custom := false
	for _, limit := range limits {
		if limit.VolumeTypeID == selected.ID && limit.AllowUserQoS && limit.FullQoSDiskType {
			custom = true
			break
		}
	}
	config.ID = types.StringValue(selected.ID)
	config.Name = types.StringValue(selected.Name)
	config.Description = types.StringValue(selected.Description)
	config.IsPublic = types.BoolValue(selected.IsPublic)
	config.CustomIOPS = types.BoolValue(custom)
	if selected.ExtraSpecs == nil {
		selected.ExtraSpecs = map[string]string{}
	}
	var diags diag.Diagnostics
	config.ExtraSpecs, diags = types.MapValueFrom(ctx, types.StringType, selected.ExtraSpecs)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
