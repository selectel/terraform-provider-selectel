package blockstorage

import (
	"context"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

type volumeResource struct {
	connection
	pollInterval time.Duration
}

func NewVolumeResource() resource.Resource { return &volumeResource{pollInterval: 3 * time.Second} }

var (
	volumeDocs          = schemadoc.Resource{Name: "network volume", Location: volumeLocation}
	volumeLimitsConcept = schemadoc.Link{
		Title: "network volume limits",
		URL:   "https://docs.selectel.ru/en/cloud-servers/volumes/about-network-volumes/#network-volume-limits",
	}

	_ resource.Resource                   = (*volumeResource)(nil)
	_ resource.ResourceWithConfigure      = (*volumeResource)(nil)
	_ resource.ResourceWithImportState    = (*volumeResource)(nil)
	_ resource.ResourceWithValidateConfig = (*volumeResource)(nil)
	_ resource.ResourceWithModifyPlan     = (*volumeResource)(nil)
)

type volumeModel struct {
	ID               types.String   `tfsdk:"id"`
	ProjectID        types.String   `tfsdk:"project_id"`
	Region           types.String   `tfsdk:"region"`
	Size             types.Int64    `tfsdk:"size"`
	Name             types.String   `tfsdk:"name"`
	Description      types.String   `tfsdk:"description"`
	AvailabilityZone types.String   `tfsdk:"availability_zone"`
	VolumeType       types.String   `tfsdk:"volume_type"`
	SnapshotID       types.String   `tfsdk:"snapshot_id"`
	SourceVolumeID   types.String   `tfsdk:"source_volume_id"`
	ImageID          types.String   `tfsdk:"image_id"`
	BackupID         types.String   `tfsdk:"backup_id"`
	Metadata         types.Map      `tfsdk:"metadata"`
	ServiceMetadata  types.Map      `tfsdk:"service_metadata"`
	Attachments      types.Set      `tfsdk:"attachments"`
	OnlineResize     types.Bool     `tfsdk:"enable_online_resize"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

var attachmentTypes = map[string]attr.Type{
	"id":          types.StringType,
	"instance_id": types.StringType,
	"device":      types.StringType,
}

// volumeSources are mutually exclusive ways to create a volume from existing data.
var volumeSources = []string{"snapshot_id", "source_volume_id", "image_id", "backup_id"}

func otherSources(name string) []string {
	return slices.DeleteFunc(slices.Clone(volumeSources), func(s string) bool { return s == name })
}

// sourceAttribute makes a one of the mutually exclusive volumeSources.
func sourceAttribute(name string, a schema.StringAttribute, field schemadoc.Field) schema.StringAttribute {
	a.Optional = true
	a.Validators = append(a.Validators, schemadoc.ConflictsWith(otherSources(name)...))
	a.PlanModifiers = append(a.PlanModifiers, schemadoc.RequiresReplace())

	return volumeDocs.String(a, field)
}

// observedSource is a source that the API reports back, so import restores it.
func observedSource(validators ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:      true,
		Validators:    append([]validator.String{schemadoc.UUID()}, validators...),
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_volume_v3"
}

func (r *volumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(r.configure(req.ProviderData)...)
}

func (r *volumeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages a network volume in Cloud Servers using the v3 public API. The volume size can only increase. Detach the volume before deletion. For more information about network volumes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/volumes/about-network-volumes/).\n\nTo create a volume from an existing source, set at most one of `snapshot_id`, `source_volume_id`, `image_id`, or `backup_id`.",
		Attributes: map[string]schema.Attribute{
			"id":         volumeDocs.ID(),
			"project_id": volumeDocs.ProjectID(),
			"region":     volumeDocs.Region(),
			"size": volumeDocs.Int64(schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				Validators:    []validator.Int64{schemadoc.Between(1, 2147483647)},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Size of the network volume in GB.",
				Behavior:   "This argument is required unless `snapshot_id`, `source_volume_id`, or `backup_id` is set. It is also required with `image_id`. Only increases are supported. To increase the size of a network volume attached to a cloud server, set `enable_online_resize` to `true`.",
				Concept:    volumeLimitsConcept,
			}),
			"name": volumeDocs.String(schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Default:    schemadoc.StaticString(""),
				Validators: []validator.String{schemadoc.LengthAtMost(255), schemadoc.Trimmed()},
			}, schemadoc.Field{
				Definition: "Name of the network volume.",
				Behavior:   "With `backup_id`, this argument is required and cannot be empty.",
			}),
			"description": volumeDocs.String(schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Default:    schemadoc.StaticString(""),
				Validators: []validator.String{schemadoc.LengthAtMost(255), schemadoc.Trimmed()},
			}, schemadoc.Field{
				Definition: "Description of the network volume.",
			}),
			"availability_zone": volumeDocs.String(schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{schemadoc.NonEmptyTrimmed()},
				// UseStateForUnknown goes first: any change marks an omitted computed
				// value unknown, and RequiresReplace would treat that as a new value.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					schemadoc.RequiresReplace(),
				},
			}, schemadoc.Field{
				Definition: "Pool segment, for example, `ru-6a`.",
				Behavior:   "The API selects a segment when omitted.",
				Concept:    schemadoc.LocationLink("available pool segments", volumeLocation),
			}),
			"volume_type": volumeDocs.String(schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{schemadoc.NonEmptyTrimmed()},
				// UseStateForUnknown goes first: any change marks an omitted computed
				// value unknown, and RequiresReplace would treat that as a new value.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					schemadoc.RequiresReplace(),
				},
			}, schemadoc.Field{
				Definition: "Name or UUID of the volume type.",
				Behavior:   volumeTypeNameDescription + " This argument is required with `snapshot_id`. In other cases, the API chooses a type when omitted.",
				Source:     schemadoc.FromDataSource("selectel_compute_volume_type_v3", "volume type name"),
				Concept:    volumeTypesConcept,
			}),
			// The API rejects a volume from a snapshot without an explicit type.
			"snapshot_id": sourceAttribute("snapshot_id", observedSource(schemadoc.AlsoRequires("volume_type")), schemadoc.Field{
				Definition: "Unique identifier of the source snapshot.",
				Behavior:   "If `size` is omitted, the API uses the snapshot size. Import restores this argument.",
				Source:     schemadoc.FromDataSource("selectel_compute_snapshot_v3", "snapshot ID"),
			}),
			"source_volume_id": sourceAttribute("source_volume_id", observedSource(), schemadoc.Field{
				Definition: "Unique identifier of the source network volume.",
				Behavior:   "If `size` is omitted, the API uses the source volume size. Import restores this argument.",
				Source:     schemadoc.FromResource("selectel_compute_volume_v3", "source volume ID"),
			}),
			// A restore into a volume without a name copies the name and description of
			// the backed-up volume, which would not match the plan.
			"backup_id": sourceAttribute("backup_id", schema.StringAttribute{
				Validators: []validator.String{schemadoc.UUID(), schemadoc.AlsoRequires("name")},
			}, schemadoc.Field{
				Definition: "Unique identifier of the backup to restore.",
				Behavior:   "If `size` is omitted, the API uses the backup size. Import does not restore this argument.",
				Source:     schemadoc.FromOpenStackCLI("backup ID", "openstack volume backup list", "ID"),
			}),
			"image_id": sourceAttribute("image_id", schema.StringAttribute{
				Validators: []validator.String{schemadoc.NonEmptyTrimmed()},
			}, schemadoc.Field{
				Definition: "UUID, name, or API URL of the source image.",
				Behavior:   "This argument requires `size`. Import does not restore this argument. Learn more about the [openstack_images_image_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/images_image_v2) resource in the official OpenStack documentation.",
			}),
			"metadata": volumeDocs.Map(schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			}, schemadoc.Field{
				Definition: "Key-value pairs of the network volume.",
				// The list of special keys stays last: render appends nothing after Behavior
				// here, because the field has no source, default or concept link.
				Behavior: "Each key and value is limited to 255 characters. Keys can contain ASCII letters, digits, underscores, colons, periods, spaces, slashes, and hyphens. Metadata added by the API is available in `service_metadata`. Import does not restore this argument. Special keys:" +
					"\n  - `total_iops_sec` — custom IOPS as a decimal string, for volume types that support it. You can change the value but cannot remove it. Learn more about available values in [" + volumeLimitsConcept.Title + "](" + volumeLimitsConcept.URL + ")." +
					"\n  - `total_bytes_sec` — managed by the API and cannot be set." +
					"\n  - `" + createTokenKey + "` — reserved by the provider and cannot be set.",
			}),
			"service_metadata": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "All metadata returned by the API, excluding the provider creation token. The map includes server-managed QoS values.",
			},
			"enable_online_resize": volumeDocs.Bool(schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  schemadoc.StaticBool(false),
			}, schemadoc.Field{
				Definition: "Specifies if Terraform can increase the size of a network volume that is attached to a cloud server.",
				Behavior:   "The provider does not expand partitions or file systems.",
				Concept: schemadoc.Link{
					Topic:  "extending attached volumes",
					Vendor: "OpenStack",
					URL:    "https://docs.openstack.org/cinder/latest/cli/cli-manage-volumes.html#extend-attached-volume",
				},
			}),
			"attachments": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Read-only volume attachments. Detach all attachments before deletion.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "Unique identifier of the network volume.",
					},
					"instance_id": schema.StringAttribute{Computed: true, Description: "Unique identifier of the attached cloud server."},
					"device":      schema.StringAttribute{Computed: true, Description: "Device name on the cloud server."},
				}},
			},
		},
		Blocks: map[string]schema.Block{"timeouts": volumeDocs.Timeouts(ctx, volumeTimeouts)},
	}
}
