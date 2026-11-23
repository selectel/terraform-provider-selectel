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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	// The window end follows these changes and cannot be kept from the state.
	if !plan.KubeVersion.Equal(state.KubeVersion) || !plan.MaintenanceWindowStart.Equal(state.MaintenanceWindowStart) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("maintenance_window_end"), types.StringUnknown())...)
	}

	x509Plan, x509Known := "", false
	planKubeOptions := plan.KubernetesOptions
	if planKubeOptions.IsNull() {
		x509Known = true
	} else if !planKubeOptions.IsUnknown() {
		var options clusterKubeOptionsModel
		if !planKubeOptions.As(ctx, &options, basetypes.ObjectAsOptions{}).HasError() {
			x509Plan = options.X509CACertificates.ValueString()
			// An unknown planned value reads as "", which would fail the plan.
			x509Known = !options.X509CACertificates.IsUnknown()
		}
	}
	if x509Known && x509Plan == "" {
		x509State := ""
		stateKubeOptions := state.KubernetesOptions
		if !stateKubeOptions.IsNull() && !stateKubeOptions.IsUnknown() {
			var options clusterKubeOptionsModel
			if !stateKubeOptions.As(ctx, &options, basetypes.ObjectAsOptions{}).HasError() {
				x509State = options.X509CACertificates.ValueString()
			}
		}
		if x509State != "" {
			resp.Diagnostics.AddAttributeError(path.Root("kubernetes_options").AtName("x509_ca_certificates"),
				"X509 CA certificates cannot be removed",
				"The API keeps the X509 CA certificates of the cluster, an empty value does not remove them. "+
					"New certificates can be set instead.")
		}
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

	enabled, known, diags := oidcEnabled(ctx, req, oidc)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !known {
		if req.ConfigValue.IsNull() {
			resp.PlanValue = types.StringUnknown()
		}

		return
	}

	if req.ConfigValue.IsUnknown() {
		// The plan must keep an unknown configured value; the apply-time plan sees the known one.
		return
	}

	if enabled {
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

// auditSecretPlanModifier plans audit_logs.secret_name the way the API stores
// it: empty when audit logs are disabled, the state name or `mks-audit-logs`
// when they are enabled.
type auditSecretPlanModifier struct {
	schemadoc.Facts
}

func (m auditSecretPlanModifier) Description(_ context.Context) string {
	return "Plans the value the API stores for the audit logs secret."
}

func (m auditSecretPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m auditSecretPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var auditLogs types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath(), &auditLogs)...)
	if resp.Diagnostics.HasError() || auditLogs.IsNull() || auditLogs.IsUnknown() {
		// An omitted audit_logs object keeps the state, see UseStateForUnknown.
		return
	}
	var planned types.Object
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, req.Path.ParentPath(), &planned)...)
	if resp.Diagnostics.HasError() || planned.IsNull() || planned.IsUnknown() {
		return
	}
	enabled := planned.Attributes()["enabled"].(types.Bool)
	if enabled.IsUnknown() {
		if req.ConfigValue.IsNull() || req.ConfigValue.ValueString() == "" {
			resp.PlanValue = types.StringUnknown()
		}

		return
	}
	if enabled.ValueBool() {
		if req.ConfigValue.IsUnknown() {
			// The apply-time plan sees the known value.
			return
		}
		if req.ConfigValue.IsNull() || req.ConfigValue.ValueString() == "" {
			if !req.StateValue.IsNull() && req.StateValue.ValueString() != "" {
				// The API keeps the name of a cluster that already collects logs.
				resp.PlanValue = req.StateValue

				return
			}
			resp.PlanValue = types.StringValue("mks-audit-logs")
		}

		return
	}
	if !req.ConfigValue.IsNull() && req.ConfigValue.ValueString() != "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Audit logs secret name set while audit logs are disabled",
			"The API rejects it: audit_logs.secret_name cannot be non-empty while audit_logs.enabled is false.")

		return
	}
	resp.PlanValue = types.StringValue("")
}

// oidcEnabled is the planned oidc.enabled: the configured value, otherwise
// the state one, otherwise false, which create sends.
func oidcEnabled(ctx context.Context, req planmodifier.StringRequest, oidc types.Object) (bool, bool, diag.Diagnostics) {
	enabled, _ := oidc.Attributes()["enabled"].(types.Bool)
	if enabled.IsUnknown() {
		return false, false, nil
	}
	if !enabled.IsNull() {
		return enabled.ValueBool(), true, nil
	}
	if req.State.Raw.IsNull() {
		return false, true, nil
	}

	var stateEnabled types.Bool
	diags := req.State.GetAttribute(ctx, req.Path.ParentPath().AtName("enabled"), &stateEnabled)
	if stateEnabled.IsNull() || stateEnabled.IsUnknown() {
		return false, false, diags
	}

	return stateEnabled.ValueBool(), true, diags
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
