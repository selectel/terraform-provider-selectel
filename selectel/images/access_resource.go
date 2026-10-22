package images

import (
	"context"
	"fmt"

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

// accessResource grants a project access to an image. It runs in the project
// that owns the image.
type accessResource struct{ connection }

func NewAccessResource() resource.Resource { return &accessResource{} }

var (
	accessDocs = schemadoc.Resource{Name: "image access"}

	_ resource.Resource                = (*accessResource)(nil)
	_ resource.ResourceWithConfigure   = (*accessResource)(nil)
	_ resource.ResourceWithImportState = (*accessResource)(nil)
)

type accessModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	Region    types.String `tfsdk:"region"`
	membership
}

func (r *accessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_image_access_v2"
}

func (r *accessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(r.configure(req.ProviderData)...)
}

func (r *accessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages access of another project to a shared image in Cloud Servers using the v2 public API. Access works only within a pool: the receiving project can use the image only in the pool where the image is located. The image must have `visibility = \"shared\"`. Granting access requires a role that allows managing image access in the project. For more information about access to images and the roles, see the [official Selectel documentation](" + accessRolesDocs + ").\n\n" +
			"The resource is managed in the project that owns the image. The receiving project accepts or rejects access with the [selectel_compute_image_access_accept_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/compute_image_access_accept_v2) resource. " +
			"If the image visibility changes from `shared`, the API keeps the access but hides it and stops applying it: Terraform removes the resource from the state, and the access applies again when the image is shared again. " +
			"To revoke access, destroy the resource before you change the image visibility. To manage hidden access after the image is shared again, import it: creating it fails because the access already exists.",
		Attributes: map[string]schema.Attribute{
			"id": accessDocs.String(schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the image access in the `<image_id>/<member_id>` format.",
			}),
			"project_id": accessDocs.ProjectID(),
			"region":     accessDocs.Region(),
			"image_id": accessDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{schemadoc.UUID()},
				PlanModifiers: []planmodifier.String{schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the shared image.",
				Source:     schemadoc.FromResource("selectel_compute_image_v2", "image ID"),
			}),
			"member_id": accessDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{schemadoc.NonEmptyTrimmed()},
				PlanModifiers: []planmodifier.String{schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the project that receives access.",
				Source:     schemadoc.FromResource("selectel_vpc_project_v2", "project ID"),
			}),
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the access: `pending`, `accepted`, or `rejected`. A new access is `pending`.",
			},
			"created_at": computed("Time when the access was granted in the RFC 3339 format."),
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Time when the access was last updated in the RFC 3339 format.",
			},
		},
	}
}

func (r *accessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accessModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	client, err := r.client(ctx, plan.ProjectID.ValueString(), plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to the Image service", err.Error())
		return
	}
	imageID, memberID := plan.ImageID.ValueString(), plan.MemberID.ValueString()
	created, _, err := member.Create(ctx, client, imageID, memberID)
	if sdk.IsKind(err, sdk.KindConflict) {
		resp.Diagnostics.AddError(
			"Image access already exists",
			fmt.Sprintf("Import the access with the ID %s to manage it.", membershipID(imageID, memberID)),
		)

		return
	}
	// The membership is addressed by its image and project, so a request
	// without a definite answer is resolved by reading the membership.
	if createOutcomeUncertain(err) {
		if found, _, getErr := member.Get(ctx, client, imageID, memberID); getErr == nil {
			created, err = found, nil
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Grant image access", err.Error())
		return
	}
	plan.observe(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *accessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accessModel
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
	// The API also answers 404 while the image is not shared: it keeps the
	// membership but hides it. The resource description explains this case.
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

// Update is never called: every argument requires replacement.
func (r *accessResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {}

func (r *accessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accessModel
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
	_, err = member.Delete(ctx, client, state.ImageID.ValueString(), state.MemberID.ValueString())
	switch {
	case err == nil || sdk.IsKind(err, sdk.KindNotFound):
	case sdk.IsKind(err, sdk.KindForbidden):
		resp.Diagnostics.AddError(
			"Revoke image access",
			err.Error()+"\n\nThe API rejects the request if the image is not shared or the project role does not allow managing image access. "+
				"To revoke access, set visibility = \"shared\" on the image, apply, and destroy the access again.",
		)
	default:
		resp.Diagnostics.AddError("Revoke image access", err.Error())
	}
}

func (r *accessResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	imageID, memberID, err := parseMembershipID(req.ID, true)
	if err != nil {
		resp.Diagnostics.AddError("Invalid image access ID", err.Error())
		return
	}
	project, region, diags := r.importScope("")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := accessModel{ProjectID: types.StringValue(project), Region: types.StringValue(region)}
	state.ID = types.StringValue(membershipID(imageID, memberID))
	state.ImageID = types.StringValue(imageID)
	state.MemberID = types.StringValue(memberID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
