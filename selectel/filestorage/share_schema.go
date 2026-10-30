package filestorage

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

type shareResource struct {
	connection
	pollInterval time.Duration
	// recoveryWindow bounds the search for a share whose create request got no
	// definite answer: the fork may list it some time after the request.
	recoveryWindow time.Duration
}

func NewShareResource() resource.Resource {
	return &shareResource{pollInterval: 5 * time.Second, recoveryWindow: 2 * time.Minute}
}

var (
	_ resource.Resource                   = (*shareResource)(nil)
	_ resource.ResourceWithConfigure      = (*shareResource)(nil)
	_ resource.ResourceWithValidateConfig = (*shareResource)(nil)
	_ resource.ResourceWithModifyPlan     = (*shareResource)(nil)
	_ resource.ResourceWithImportState    = (*shareResource)(nil)
)

type shareModel struct {
	ID               types.String   `tfsdk:"id"`
	ProjectID        types.String   `tfsdk:"project_id"`
	Region           types.String   `tfsdk:"region"`
	ShareProto       types.String   `tfsdk:"share_proto"`
	Size             types.Int64    `tfsdk:"size"`
	ShareType        types.String   `tfsdk:"share_type"`
	NetworkID        types.String   `tfsdk:"network_id"`
	SubnetID         types.String   `tfsdk:"subnet_id"`
	IP               types.String   `tfsdk:"ip"`
	Name             types.String   `tfsdk:"name"`
	Description      types.String   `tfsdk:"description"`
	Metadata         types.Map      `tfsdk:"metadata"`
	AllMetadata      types.Map      `tfsdk:"all_metadata"`
	Status           types.String   `tfsdk:"status"`
	AvailabilityZone types.String   `tfsdk:"availability_zone"`
	ExportLocations  types.List     `tfsdk:"export_locations"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

var exportLocationTypes = map[string]attr.Type{
	"path":      types.StringType,
	"preferred": types.BoolType,
}

// aboutFileStorage is the product overview of File Storage.
const (
	aboutFileStorage = "https://docs.selectel.ru/en/file-storage/about/about-file-storage/"
	accessRulesDocs  = "https://docs.selectel.ru/en/file-storage/manage/configure-access-by-ips/"
)

var shareTypesConcept = schemadoc.Link{
	Title: "File storage types",
	URL:   aboutFileStorage + "#file-storage-types",
}

func (r *shareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filestorage_share_v2"
}

func (r *shareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(r.configure(req.ProviderData)...)
}

// replaceString is an argument that the API cannot change on an existing share.
func replaceString(validators ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		Required:      true,
		Validators:    validators,
		PlanModifiers: []planmodifier.String{schemadoc.RequiresReplace()},
	}
}

func (r *shareResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages a file storage in File Storage using the v2 public API. The file storage is placed in a private network and is created without access rules, so no client can mount it until you add `selectel_filestorage_share_access_v2`. The size can only increase. For more information about file storage, see the [official Selectel documentation](" + aboutFileStorage + ").",
		Attributes: map[string]schema.Attribute{
			"id":         shareDocs.ID(),
			"project_id": shareDocs.ProjectID(),
			"region":     shareDocs.Region(),
			"share_proto": shareDocs.String(replaceString(schemadoc.OneOf(protoNFS, protoCIFS)), schemadoc.Field{
				Definition: "File storage protocol.",
			}),
			"size": shareDocs.Int64(schema.Int64Attribute{
				Required:   true,
				Validators: []validator.Int64{schemadoc.AtLeast(shareMinSize)},
			}, schemadoc.Field{
				Definition: "Size of the file storage in GB.",
				Behavior:   "Only increases are supported. Unmount the file storage from all clients before increasing the size and mount it again after.",
			}),
			"share_type": shareDocs.String(replaceString(schemadoc.NonEmptyTrimmed()), schemadoc.Field{
				Definition: "Name of the file storage type.",
				Behavior: "The name format is `<type>.<pool_segment>`, where `<type>` is `basic` for HDD Basic, `universal` for SSD Universal, or `fast` for SSD Fast, " +
					"for example, `basic.ru-6a` for HDD Basic in the `ru-6a` pool segment. The type also sets the pool segment of the file storage. " +
					"Not every type is available in every pool segment: see [File Storage availability](https://docs.selectel.ru/en/infrastructure/product-availability-by-location/#file-storage). A UUID is rejected: use the type name.",
				Source:  schemadoc.FromDataSource("selectel_filestorage_availability_zones_v2", "pool segments"),
				Concept: shareTypesConcept,
			}),
			"network_id": shareDocs.String(replaceString(schemadoc.UUID()), schemadoc.Field{
				Definition: "Unique identifier of the private network of the file storage.",
				Source:     schemadoc.FromResource("selectel_vpc_network_v2", "network ID"),
			}),
			"subnet_id": shareDocs.String(replaceString(schemadoc.UUID()), schemadoc.Field{
				Definition: "Unique identifier of the private subnet of the file storage.",
				Behavior: "A subnet of a cloud network gives access to cloud servers and Managed Kubernetes clusters in the same pool. " +
					"To connect dedicated servers or servers in other pools, use a subnet connected to a global router: see [Connect a File Storage](https://docs.selectel.ru/en/file-storage/add/).",
				Source: schemadoc.FromResource("selectel_vpc_network_subnet_v2", "subnet ID"),
			}),
			"ip": shareDocs.String(schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{ipv4Validator()},
				// UseStateForUnknown goes first: an omitted ip keeps the observed
				// address, and only a configured different address replaces the share.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					schemadoc.RequiresReplaceIf(replacesKnownIP, ""),
				},
			}, schemadoc.Field{
				Definition: "IP address of the file storage in the subnet.",
				Behavior:   "When omitted, the first free address of the subnet is used.",
			}),
			"name": shareDocs.String(schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{schemadoc.LengthAtMost(255)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Name of the file storage.",
				Behavior:   "When omitted, the resource keeps the current value, so an imported file storage keeps the value set in the control panel, and the plan does not show changes made outside Terraform. To clear the value, set an empty string.",
			}),
			"description": shareDocs.String(schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{schemadoc.LengthAtMost(255)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Description of the file storage.",
				Behavior:   "When omitted, the resource keeps the current value, so an imported file storage keeps the value set in the control panel, and the plan does not show changes made outside Terraform. To clear the value, set an empty string.",
			}),
			"metadata": shareDocs.Map(schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			}, schemadoc.Field{
				Definition: "Key-value pairs of the file storage.",
				Behavior: "The resource manages only the keys set here; other keys are available in `all_metadata`. Import does not restore this argument. " +
					"These keys are reserved by the service or the provider and cannot be set: " + schemadoc.CodeList(reservedMetadataKeys) + ".",
			}),
			"all_metadata": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "All metadata of the file storage, including the keys added by the service.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the file storage.",
			},
			"availability_zone": schema.StringAttribute{
				Computed:      true,
				Description:   "Pool segment of the file storage.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"export_locations": schema.ListNestedAttribute{
				Computed:      true,
				Description:   "Paths to mount the file storage.",
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Computed:    true,
						Description: "Mount path, for example, `192.168.199.5:/shares/share-<id>` for NFS or `\\\\192.168.199.5\\share-<id>` for CIFS.",
					},
					"preferred": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the API recommends this path.",
					},
				}},
			},
		},
		Blocks: map[string]schema.Block{"timeouts": shareDocs.Timeouts(ctx, shareTimeouts)},
	}
}
