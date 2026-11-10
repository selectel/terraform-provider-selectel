package filestorage

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

// recreateHint is the way around the in-place restrictions of ModifyPlan.
const recreateHint = "To recreate the file storage instead, run terraform taint for this resource and apply."

func (r *shareResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config shareModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// The fork resolves the type by name only; a UUID from the type data source
	// would fail at create time with a misleading answer.
	if !config.ShareType.IsUnknown() && uuidPattern.MatchString(config.ShareType.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("share_type"),
			"Share type UUID is not supported",
			"Set the type name in the format <type>.<pool_segment>, for example, basic.ru-6a.",
		)
	}
	resp.Diagnostics.Append(reservedMetadataErrors(config.Metadata)...)
}

// reservedMetadataErrors rejects the keys of reservedMetadataKeys. An unknown
// map has no elements at validation, so Create and Update check it again.
func reservedMetadataErrors(metadata types.Map) diag.Diagnostics {
	var diags diag.Diagnostics
	for key := range metadata.Elements() {
		if !slices.Contains(reservedMetadataKeys, key) {
			continue
		}
		detail := "These keys are managed by the file storage service or the provider: " +
			schemadoc.CodeList(reservedMetadataKeys) + "."
		switch key {
		case "ip":
			detail += " To set the address of the file storage, use the ip argument."
		case "access_list_allow_all":
			detail += ` To allow all addresses of the private subnet, use selectel_filestorage_share_access_v2 with access_to = "0.0.0.0/0".`
		}
		diags.AddAttributeError(path.Root("metadata").AtMapKey(key), "Reserved metadata key", detail)
	}

	return diags
}

func (r *shareResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var prior, plan shareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// A replacement may use a smaller share, but only one caused by the
	// configuration is visible here: Terraform decides on -replace after
	// planning, so the error points to taint instead.
	if hasCreationChange(prior, plan) {
		return
	}
	if !plan.Size.IsUnknown() && plan.Size.ValueInt64() < prior.Size.ValueInt64() {
		resp.Diagnostics.AddAttributeError(
			path.Root("size"),
			"File storage shrinking is unsupported",
			fmt.Sprintf("Current size is %d GB. Set size to this value or greater. %s", prior.Size.ValueInt64(), recreateHint),
		)
	}
}

// hasCreationChange reports whether the plan replaces the share. It repeats
// the RequiresReplace rules of the schema, so change both together: Framework
// runs ModifyPlan with an empty RequiresReplace list, and the decisions of the
// attribute plan modifiers are not available here.
func hasCreationChange(prior, plan shareModel) bool {
	for _, change := range []struct{ before, after types.String }{
		{prior.ProjectID, plan.ProjectID},
		{prior.Region, plan.Region},
		{prior.ShareProto, plan.ShareProto},
		{prior.ShareType, plan.ShareType},
		{prior.NetworkID, plan.NetworkID},
		{prior.SubnetID, plan.SubnetID},
	} {
		if !change.after.Equal(change.before) {
			return true
		}
	}

	return ipReplaced(prior.IP, plan.IP)
}

// ipReplaced reports whether a configured address differs from the observed
// one. An empty prior address is unknown, as after an import whose mount paths
// contain no IPv4 address, so the configured address is stored instead.
func ipReplaced(prior, planned types.String) bool {
	return prior.ValueString() != "" && !planned.Equal(prior)
}

func replacesKnownIP(
	_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse,
) {
	resp.RequiresReplace = ipReplaced(req.StateValue, req.PlanValue)
}
