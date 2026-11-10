package filestorage

import (
	"context"
	"net/netip"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

type shareAccessResource struct {
	connection
	pollInterval   time.Duration
	recoveryWindow time.Duration
}

func NewShareAccessResource() resource.Resource {
	return &shareAccessResource{pollInterval: 3 * time.Second, recoveryWindow: 2 * time.Minute}
}

var (
	_ resource.Resource                = (*shareAccessResource)(nil)
	_ resource.ResourceWithConfigure   = (*shareAccessResource)(nil)
	_ resource.ResourceWithImportState = (*shareAccessResource)(nil)
)

var accessTimeouts = schemadoc.Timeouts{
	Create: 10 * time.Minute,
	Delete: 10 * time.Minute,
}

const (
	levelRW = "rw"
	levelRO = "ro"
)

type shareAccessModel struct {
	ID          types.String   `tfsdk:"id"`
	ProjectID   types.String   `tfsdk:"project_id"`
	Region      types.String   `tfsdk:"region"`
	ShareID     types.String   `tfsdk:"share_id"`
	AccessTo    types.String   `tfsdk:"access_to"`
	AccessLevel types.String   `tfsdk:"access_level"`
	State       types.String   `tfsdk:"state"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (r *shareAccessResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_filestorage_share_access_v2"
}

func (r *shareAccessResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	resp.Diagnostics.Append(r.configure(req.ProviderData)...)
}

func (r *shareAccessResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{schemadoc.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages an access rule of a file storage for an IP address or a subnet in File Storage using the v2 public API. Any change replaces the rule. " +
			"For an NFS file storage, the service applies the new list of rules within 15 minutes after a rule is added or deleted. " +
			"For more information about access rules, see the [official Selectel documentation](" + accessRulesDocs + ").",
		Attributes: map[string]schema.Attribute{
			"id": accessDocs.String(schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the access rule.",
				Behavior:   "Import uses it together with the file storage ID.",
				Source:     schemadoc.FromOpenStackCLI("access rule ID", "openstack share access list <share_id>", "ID"),
			}),
			"project_id": accessDocs.ProjectID(),
			"region": accessDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{schemadoc.NonEmptyTrimmed()},
				PlanModifiers: replace,
			}, schemadoc.Field{
				Definition: "Pool of the file storage, for example, `ru-6`.",
				Behavior:   "It must match the pool of the file storage.",
			}),
			"share_id": accessDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{schemadoc.UUID()},
				PlanModifiers: replace,
			}, schemadoc.Field{
				Definition: "Unique identifier of the file storage.",
				Source:     schemadoc.FromResource("selectel_filestorage_share_v2", "file storage ID"),
			}),
			"access_to": accessDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{accessTargetValidator()},
				PlanModifiers: replace,
			}, schemadoc.Field{
				Definition: "IP address or subnet that gets access, for example, `192.168.199.10` or `192.168.199.0/24`.",
				Behavior: "`0.0.0.0/0` allows all addresses of the private subnet of the file storage and overrides narrower rules; no metadata key is needed for it. " +
					"Addresses of other private subnets are allowed when they are routed to the subnet of the file storage, for example, through a global router.",
			}),
			"access_level": accessDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{schemadoc.OneOf(levelRW, levelRO)},
				PlanModifiers: replace,
			}, schemadoc.Field{
				Definition: "Access level: `rw` for read and write, `ro` for read only.",
				Behavior:   "A CIFS file storage supports only `rw`.",
			}),
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "State of the access rule, for example, `active`.",
			},
		},
		Blocks: map[string]schema.Block{"timeouts": accessDocs.Timeouts(ctx, accessTimeouts)},
	}
}

// accessTargetValidator accepts an IPv4 address or an IPv4 CIDR without host
// bits, which the API would reject.
func accessTargetValidator() validator.String {
	return schemadoc.DescribeString(accessTarget{},
		"The format is an IPv4 address or an IPv4 CIDR without host bits, for example, `192.168.199.0/24`.")
}

type accessTarget struct{}

func (accessTarget) Description(context.Context) string {
	return "must be an IPv4 address or an IPv4 CIDR without host bits"
}

func (v accessTarget) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func (v accessTarget) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !validAccessTarget(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid access target",
			"Value "+req.ConfigValue.String()+" "+v.Description(ctx)+".")
	}
}

func validAccessTarget(value string) bool {
	if address, err := netip.ParseAddr(value); err == nil {
		return address.Is4()
	}
	prefix, err := netip.ParsePrefix(value)

	return err == nil && prefix.Addr().Is4() && prefix.Masked() == prefix
}
