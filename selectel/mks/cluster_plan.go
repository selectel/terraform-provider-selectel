package mks

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mksclient "github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

func (r *clusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var clusterType, cniType, workersType types.String
	var cloudSubnetCIDR types.String
	var autoUpgrade, autorepair types.Bool
	var cilium types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("cluster_type"), &clusterType)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("workers_type"), &workersType)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("cloud_subnet_cidr"), &cloudSubnetCIDR)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("enable_patch_version_auto_upgrade"), &autoUpgrade)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("enable_autorepair"), &autorepair)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("cni_type"), &cniType)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("cni_cilium_settings"), &cilium)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API stores Cilium settings only for a Cilium cluster.
	if !cilium.IsNull() && !cilium.IsUnknown() && !cniType.IsUnknown() &&
		cniType.ValueString() != string(mksclient.ClusterCniTypeCILIUM) {
		resp.Diagnostics.AddAttributeError(path.Root("cni_cilium_settings"),
			"Cilium settings need the Cilium CNI",
			"Set cni_type to CILIUM or remove cni_cilium_settings.")
	}

	// The API rejects it.
	if clusterType.ValueString() == string(mksclient.ClusterCreateStructClusterTypeBASIC) && autoUpgrade.ValueBool() {
		resp.Diagnostics.AddAttributeError(path.Root("enable_patch_version_auto_upgrade"),
			"Patch version auto-upgrade is not available for BASIC clusters",
			"Set enable_patch_version_auto_upgrade to false or omit it when cluster_type is BASIC.")
	}

	// The API ignores the CIDR for a non-dedicated cluster.
	if !cloudSubnetCIDR.IsNull() && !cloudSubnetCIDR.IsUnknown() && workersType.ValueString() == "CLOUD" {
		resp.Diagnostics.AddAttributeError(path.Root("cloud_subnet_cidr"),
			"Cloud subnet CIDR is not available for cloud clusters",
			"Remove cloud_subnet_cidr or set workers_type to DEDICATED.")
	}

	// The API always disables auto-repair for a dedicated cluster.
	if workersType.ValueString() == "DEDICATED" && autorepair.ValueBool() {
		resp.Diagnostics.AddAttributeError(path.Root("enable_autorepair"),
			"Auto-repair is not available for dedicated clusters",
			"The API always disables enable_autorepair for a DEDICATED cluster, so setting it to true ends in an inconsistent result.")
	}
}

func (r *clusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state, plan clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.KubeVersion.IsUnknown() && !plan.KubeVersion.Equal(state.KubeVersion) {
		order, err := compareKubeVersions(plan.KubeVersion.ValueString(), state.KubeVersion.ValueString())
		if err == nil && order < 0 {
			// Like the v1 diff suppress: a lower configured version, left behind by patch auto-upgrade, changes nothing.
			plan.KubeVersion = state.KubeVersion
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("kube_version"), state.KubeVersion)...)
		} else {
			err = validateKubeVersionChange(state.KubeVersion.ValueString(), plan.KubeVersion.ValueString())
			if err != nil {
				resp.Diagnostics.AddAttributeError(path.Root("kube_version"), "Invalid Kubernetes version change", err.Error())

				return
			}
		}
	}

	// Any update can change the status: a task-creating PATCH moves the cluster to PENDING_UPGRADE_CLUSTER_CONFIG, a finished task leaves it ACTIVE or MAINTENANCE.
	if !resp.Plan.Raw.Equal(req.State.Raw) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("status"), types.StringUnknown())...)
	}
	// The API derives the window end from the start alone.
	if !plan.MaintenanceWindowStart.Equal(state.MaintenanceWindowStart) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("maintenance_window_end"), types.StringUnknown())...)
	}
}

// oidcPlanModifier plans an OIDC parameter the way the API stores it: empty
// when OIDC is disabled, the API default for an omitted claim when enabled.
type oidcPlanModifier struct {
	enabledDefault string
	schemadoc.Facts
}

func (m oidcPlanModifier) Description(_ context.Context) string {
	return "Plans the value the API stores for the OIDC parameter."
}

func (m oidcPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m oidcPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var oidc types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath(), &oidc)...)
	if resp.Diagnostics.HasError() || oidc.IsNull() || oidc.IsUnknown() {
		// An omitted oidc object keeps the state, see UseStateForUnknown.
		return
	}

	enabled, diags := oidcEnabled(ctx, req, oidc)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if enabled.IsUnknown() {
		if req.ConfigValue.IsNull() {
			resp.PlanValue = types.StringUnknown()
		}

		return
	}

	if req.ConfigValue.IsUnknown() {
		// The plan must keep an unknown configured value; the apply-time plan sees the known one.
		return
	}

	if enabled.ValueBool() {
		if m.enabledDefault == "" {
			return
		}
		if req.ConfigValue.IsNull() {
			resp.PlanValue = types.StringValue(m.enabledDefault)

			return
		}
		if req.ConfigValue.ValueString() == "" {
			// The plan cannot differ from the configured ""; the API would store the default.
			name, _ := req.Path.Steps().LastStep()
			resp.Diagnostics.AddAttributeError(req.Path, "Empty OIDC claim while OIDC is enabled",
				fmt.Sprintf("%s cannot be empty while OIDC is enabled: omit it to get `%s`.", name, m.enabledDefault))
		}

		return
	}

	if !req.ConfigValue.IsNull() && req.ConfigValue.ValueString() != "" {
		resp.Diagnostics.AddAttributeError(req.Path, "OIDC parameter set while OIDC is disabled",
			fmt.Sprintf("The API rejects it: oidc parameters cannot be configured when it is disabled. "+
				"Disabling OIDC clears its settings, so remove %s.", req.Path))

		return
	}
	resp.PlanValue = types.StringValue("")
}

// oidcEnabled is the planned oidc.enabled: the configured value, otherwise
// the state one, otherwise false, which create sends.
func oidcEnabled(ctx context.Context, req planmodifier.StringRequest, oidc types.Object) (types.Bool, diag.Diagnostics) {
	enabled, _ := oidc.Attributes()["enabled"].(types.Bool)
	if !enabled.IsNull() {
		return enabled, nil
	}
	if req.State.Raw.IsNull() {
		return types.BoolValue(false), nil
	}

	var stateEnabled types.Bool
	diags := req.State.GetAttribute(ctx, req.Path.ParentPath().AtName("enabled"), &stateEnabled)
	if stateEnabled.IsNull() {
		return types.BoolUnknown(), diags
	}

	return stateEnabled, diags
}

// nameCase plans the state name for a configured name that differs only in
// letter case: the API lower-cases the name, and v1 ignored the case too.
type nameCase struct {
	schemadoc.Facts
}

func (m nameCase) Description(_ context.Context) string {
	return "A change of only the letter case keeps the state value."
}

func (m nameCase) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m nameCase) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}
	if strings.EqualFold(req.PlanValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
