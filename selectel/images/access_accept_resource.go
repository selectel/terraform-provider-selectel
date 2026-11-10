package images

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/images-go/pkg/v1"
	"github.com/selectel/images-go/pkg/v1/member"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

// accessAcceptResource sets the status of the access that a project received.
// It runs in the receiving project: only that project can change the status,
// and only the image owner can remove the access.
type accessAcceptResource struct{ connection }

func NewAccessAcceptResource() resource.Resource { return &accessAcceptResource{} }

var (
	acceptDocs = schemadoc.Resource{Name: "image access acceptance"}

	_ resource.Resource                   = (*accessAcceptResource)(nil)
	_ resource.ResourceWithConfigure      = (*accessAcceptResource)(nil)
	_ resource.ResourceWithImportState    = (*accessAcceptResource)(nil)
	_ resource.ResourceWithValidateConfig = (*accessAcceptResource)(nil)
)

type accessAcceptModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	Region    types.String `tfsdk:"region"`
	membership
}

func (r *accessAcceptResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_compute_image_access_accept_v2"
}

func (r *accessAcceptResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	resp.Diagnostics.Append(r.configure(req.ProviderData)...)
}

func (r *accessAcceptResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages the acceptance of access to a shared image in Cloud Servers using the v2 public API. Access works only within a pool: the receiving project can use the image only in the pool where the image is located. Changing the status requires a role that allows managing image access in the receiving project. For more information about access to images and the roles, see the [official Selectel documentation](" + accessRolesDocs + ").\n\n" +
			"The resource is managed in the project that received access. The image owner grants access with the [selectel_compute_image_access_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/compute_image_access_v2) resource. " +
			"Destroying the resource rejects the access instead of removing it: only the image owner can remove access. " +
			"If the image owner changes the image visibility from `shared`, the access is hidden from the receiving project: Terraform removes the resource from the state and creates it again after the image is shared again.",
		Attributes: map[string]schema.Attribute{
			"id": acceptDocs.String(schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the image access in the `<image_id>/<member_id>` format.",
			}),
			"project_id": acceptDocs.ProjectID(),
			"region":     acceptDocs.Region(),
			"image_id": acceptDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{schemadoc.UUID()},
				PlanModifiers: []planmodifier.String{schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the shared image.",
				Source: schemadoc.FromResource("selectel_compute_image_access_v2", "image ID") + " " +
					schemadoc.FromDataSource("selectel_compute_image_v2", "image ID") +
					" To find an image shared with the project, set `visibility = \"shared\"` and `member_status = \"all\"` in the data source.",
			}),
			"member_id": acceptDocs.String(schema.StringAttribute{
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
				Definition: "Unique identifier of the project that received access.",
				Behavior:   "A project can change only its own access, so the value must match `project_id`. If omitted, the provider uses `project_id`.",
			}),
			"status": acceptDocs.String(schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{schemadoc.OneOf(member.StatusAccepted, member.StatusRejected, member.StatusPending)},
			}, schemadoc.Field{
				Definition: "Status of the access.",
				Behavior:   "With `accepted`, the image appears in the default image list of the project.",
			}),
			"created_at": computed("Time when the access was granted in the RFC 3339 format."),
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Time when the access was last updated in the RFC 3339 format.",
			},
		},
	}
}

// ValidateConfig rejects the access of another project: the API lets a
// project change only its own membership.
func (r *accessAcceptResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config accessAcceptModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.MemberID.IsNull() || config.MemberID.IsUnknown() ||
		config.ProjectID.IsUnknown() {
		return
	}
	if !config.MemberID.Equal(config.ProjectID) {
		resp.Diagnostics.AddAttributeError(
			path.Root("member_id"),
			"Access of another project",
			"A project can change only its own access. Set member_id to project_id or omit it.",
		)
	}
}

func (r *accessAcceptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accessAcceptModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.MemberID.IsUnknown() {
		plan.MemberID = plan.ProjectID
	}
	resp.Diagnostics.Append(r.applyStatus(ctx, &plan)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *accessAcceptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan accessAcceptModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.applyStatus(ctx, &plan)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

// applyStatus sets the planned status of the membership and writes the
// membership that the API reports into the model. Setting a status is
// idempotent, so Create and Update are the same request.
func (r *accessAcceptResource) applyStatus(ctx context.Context, plan *accessAcceptModel) diag.Diagnostics {
	var diags diag.Diagnostics
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	client, err := r.client(ctx, plan.ProjectID.ValueString(), plan.Region.ValueString())
	if err != nil {
		diags.AddError("Connect to the Image service", err.Error())
		return diags
	}
	updated, _, err := member.Update(
		ctx, client, plan.ImageID.ValueString(), plan.MemberID.ValueString(), plan.Status.ValueString(),
	)
	if err != nil {
		diags.AddError("Set image access status", err.Error())
		return diags
	}
	plan.observe(updated)

	return diags
}

func (r *accessAcceptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accessAcceptModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	client, err := r.client(ctx, state.ProjectID.ValueString(), state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to the Image service", err.Error())
		return
	}
	current, _, err := member.Get(ctx, client, state.ImageID.ValueString(), state.MemberID.ValueString())
	// The receiving project cannot tell a removed access from a hidden one:
	// the API answers 404 in both cases. The resource description explains it.
	if sdk.IsKind(err, sdk.KindNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read image access", err.Error())
		return
	}
	state.observe(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete rejects the access: only the image owner can remove it.
func (r *accessAcceptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accessAcceptModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	client, err := r.client(ctx, state.ProjectID.ValueString(), state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to the Image service", err.Error())
		return
	}
	_, _, err = member.Update(
		ctx, client, state.ImageID.ValueString(), state.MemberID.ValueString(), member.StatusRejected,
	)
	if err != nil && !sdk.IsKind(err, sdk.KindNotFound) {
		resp.Diagnostics.AddError("Reject image access", err.Error())
	}
}

// ImportState accepts <image_id> or <image_id>/<member_id>. The member is
// the project that received access, so the full ID names the project, and the
// short one takes it from the provider configuration.
func (r *accessAcceptResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	imageID, memberID, err := parseMembershipID(req.ID, false)
	if err != nil {
		resp.Diagnostics.AddError("Invalid image access ID", err.Error())
		return
	}
	project, region, diags := r.importScope(memberID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := accessAcceptModel{ProjectID: types.StringValue(project), Region: types.StringValue(region)}
	state.ID = types.StringValue(membershipID(imageID, project))
	state.ImageID = types.StringValue(imageID)
	state.MemberID = types.StringValue(project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
