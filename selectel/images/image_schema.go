package images

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

type imageResource struct {
	connection
	pollInterval time.Duration
}

func NewImageResource() resource.Resource { return &imageResource{pollInterval: 5 * time.Second} }

var (
	imageDocs = schemadoc.Resource{Name: "image"}

	_ resource.Resource                   = (*imageResource)(nil)
	_ resource.ResourceWithConfigure      = (*imageResource)(nil)
	_ resource.ResourceWithImportState    = (*imageResource)(nil)
	_ resource.ResourceWithValidateConfig = (*imageResource)(nil)
	_ resource.ResourceWithModifyPlan     = (*imageResource)(nil)
)

type imageModel struct {
	ID                types.String   `tfsdk:"id"`
	ProjectID         types.String   `tfsdk:"project_id"`
	Region            types.String   `tfsdk:"region"`
	Name              types.String   `tfsdk:"name"`
	ContainerFormat   types.String   `tfsdk:"container_format"`
	DiskFormat        types.String   `tfsdk:"disk_format"`
	LocalFilePath     types.String   `tfsdk:"local_file_path"`
	ImageSourceURL    types.String   `tfsdk:"image_source_url"`
	MinDiskGB         types.Int64    `tfsdk:"min_disk_gb"`
	MinRAMMB          types.Int64    `tfsdk:"min_ram_mb"`
	Protected         types.Bool     `tfsdk:"protected"`
	Hidden            types.Bool     `tfsdk:"hidden"`
	Visibility        types.String   `tfsdk:"visibility"`
	Tags              types.Set      `tfsdk:"tags"`
	Properties        types.Map      `tfsdk:"properties"`
	ServiceProperties types.Map      `tfsdk:"service_properties"`
	Status            types.String   `tfsdk:"status"`
	Owner             types.String   `tfsdk:"owner"`
	SizeBytes         types.Int64    `tfsdk:"size_bytes"`
	Checksum          types.String   `tfsdk:"checksum"`
	CreatedAt         types.String   `tfsdk:"created_at"`
	UpdatedAt         types.String   `tfsdk:"updated_at"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

// sourceAttribute makes one of the two sources of the image data. The API
// does not report the source, so import cannot restore it, and an imported
// image is described by the same configuration without a replacement.
// Both sources replace the image by the same decision, sourceReplaced.
func sourceAttribute(other string, a schema.StringAttribute, field schemadoc.Field) schema.StringAttribute {
	a.Optional = true
	a.Validators = append(a.Validators, schemadoc.ConflictsWith(other))
	a.PlanModifiers = append(a.PlanModifiers, schemadoc.RequiresReplaceIf(
		func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			var prior, plan imageModel
			resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			resp.RequiresReplace = !resp.Diagnostics.HasError() && sourceReplaced(prior, plan)
		},
		"Setting a source on an image that has none, such as an imported image, or removing the source without setting another one does not replace the image.",
	))
	field.Behavior += " When Terraform creates or replaces the image, one of `local_file_path` and `image_source_url` is required. Import does not restore this argument."

	return imageDocs.String(a, field)
}

// computed is a read-only attribute that an update does not change, so the
// plan keeps the value from the state instead of showing it as unknown.
// Attributes that an update changes, such as updated_at, do not use it.
func computed(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:      true,
		Description:   description,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_image_v2"
}

func (r *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(r.configure(req.ProviderData)...)
}

func (r *imageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages an image in Cloud Servers using the v2 public API. Terraform waits until the image becomes `active`. For more information about images, see the [official Selectel documentation](" + aboutImages + ").\n\n" +
			"If the creation fails after the image is registered, for example, because the download fails or the wait times out, Terraform keeps the image in the state and marks it as tainted. The next apply replaces the image.",
		Attributes: map[string]schema.Attribute{
			"id":         imageDocs.ID(),
			"project_id": imageDocs.ProjectID(),
			"region":     imageDocs.Region(),
			"name": imageDocs.String(schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{schemadoc.LengthAtLeast(1), schemadoc.LengthAtMost(255)},
			}, schemadoc.Field{
				Definition: "Name of the image.",
			}),
			"container_format": imageDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{schemadoc.OneOf(containerFormats...)},
				PlanModifiers: []planmodifier.String{schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Container format of the image.",
				Concept:    formatsLink,
			}),
			"disk_format": imageDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{schemadoc.OneOf(diskFormats...)},
				PlanModifiers: []planmodifier.String{schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Disk format of the image.",
				Behavior:   "The API does not convert the image data, so the value must match the format of the file.",
				Concept:    formatsLink,
			}),
			"local_file_path": sourceAttribute("image_source_url", schema.StringAttribute{
				Validators: []validator.String{schemadoc.LengthAtLeast(1)},
			}, schemadoc.Field{
				Definition: "Path to the image file on the machine that runs Terraform.",
				Behavior:   "After the upload, Terraform compares the SHA-512 checksum of the file with the checksum reported by the API.",
			}),
			"image_source_url": sourceAttribute("local_file_path", schema.StringAttribute{
				Validators: []validator.String{webURL()},
			}, schemadoc.Field{
				Definition: "URL of the image file that the API downloads.",
				Behavior:   "The URL must be reachable from the Selectel network. The API stores the file as is and does not unpack archives.",
			}),
			"min_disk_gb": imageDocs.Int64(schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				Validators:    []validator.Int64{schemadoc.AtLeast(0)},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Minimum disk size in GB required to boot the image.",
				Behavior:   "If omitted, the API derives the value from the image data.",
			}),
			"min_ram_mb": imageDocs.Int64(schema.Int64Attribute{
				Optional:   true,
				Computed:   true,
				Default:    schemadoc.StaticInt64(0),
				Validators: []validator.Int64{schemadoc.AtLeast(0)},
			}, schemadoc.Field{
				Definition: "Minimum RAM in MB required to boot the image.",
			}),
			"protected": imageDocs.Bool(schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  schemadoc.StaticBool(false),
			}, schemadoc.Field{
				Definition: "Specifies if the image is protected from deletion.",
				Behavior:   "Terraform enables the protection after the image becomes `active`. To destroy or replace a protected image, first set `protected` to `false` and apply this change on its own.",
			}),
			"hidden": imageDocs.Bool(schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  schemadoc.StaticBool(false),
			}, schemadoc.Field{
				Definition: "Specifies if the image is hidden from the default image list.",
			}),
			"visibility": imageDocs.String(schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Default:    schemadoc.StaticString("private"),
				Validators: []validator.String{schemadoc.OneOf("private", "shared", "community")},
			}, schemadoc.Field{
				Definition: "Visibility of the image.",
				Behavior:   "Use `shared` to grant access to other projects with the [selectel_compute_image_access_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/compute_image_access_v2) resource. A `community` image is available to all projects in the pool.",
				Concept:    sharingLink,
			}),
			"tags": imageDocs.Set(schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators:  tagsRules(),
			}, schemadoc.Field{
				Definition: "Set of image tags.",
				Behavior:   "Tags are case-sensitive. Tags added outside of Terraform appear as changes in the plan. Import restores this argument.",
			}),
			"properties": imageDocs.Map(schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators:  []validator.Map{propertiesLimit(), propertyKeys()},
			}, schemadoc.Field{
				Definition: "Key-value pairs of custom image properties, for example, `os_distro = \"ubuntu\"`.",
				Behavior: "Keys that differ only in letter case are not allowed. " +
					"Keys of image fields, such as `name`, `min_disk`, or `os_hidden`, are not accepted: set the matching argument instead. The `" + reservedPropertyPrefix + "` prefix is reserved by the API. " +
					"Properties added by the API are available in `service_properties`. Import does not restore this argument.",
			}),
			"service_properties": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "All image properties returned by the API, including those that are not set in `properties`.",
			},
			"status":     computed("Status of the image."),
			"owner":      computed("Unique identifier of the project that owns the image."),
			"checksum":   computed("MD5 checksum of the image data."),
			"created_at": computed("Time when the image was created in the RFC 3339 format."),
			"size_bytes": schema.Int64Attribute{
				Computed:      true,
				Description:   "Size of the image data in bytes.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Time when the image was last updated in the RFC 3339 format.",
			},
		},
		Blocks: map[string]schema.Block{"timeouts": imageDocs.Timeouts(ctx, imageTimeouts)},
	}
}
