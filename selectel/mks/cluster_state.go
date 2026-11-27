package mks

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	mksclient "github.com/selectel/mks-go/v2/pkg/mksclient"
)

type clusterModel struct {
	ID                            types.String   `tfsdk:"id"`
	Name                          types.String   `tfsdk:"name"`
	Pool                          types.String   `tfsdk:"pool"`
	ProjectID                     types.String   `tfsdk:"project_id"`
	KubeVersion                   types.String   `tfsdk:"kube_version"`
	ClusterType                   types.String   `tfsdk:"cluster_type"`
	WorkersType                   types.String   `tfsdk:"workers_type"`
	NetworkID                     types.String   `tfsdk:"network_id"`
	SubnetID                      types.String   `tfsdk:"subnet_id"`
	CloudSubnetCIDR               types.String   `tfsdk:"cloud_subnet_cidr"`
	PrivateKubeAPI                types.Bool     `tfsdk:"private_kube_api"`
	MaintenanceWindowStart        types.String   `tfsdk:"maintenance_window_start"`
	MaintenanceWindowEnd          types.String   `tfsdk:"maintenance_window_end"`
	EnableAutorepair              types.Bool     `tfsdk:"enable_autorepair"`
	EnablePatchVersionAutoUpgrade types.Bool     `tfsdk:"enable_patch_version_auto_upgrade"`
	CNIType                       types.String   `tfsdk:"cni_type"`
	CNICiliumSettings             types.Object   `tfsdk:"cni_cilium_settings"`
	KubernetesOptions             types.Object   `tfsdk:"kubernetes_options"`
	Status                        types.String   `tfsdk:"status"`
	KubeAPIIP                     types.String   `tfsdk:"kube_api_ip"`
	Timeouts                      timeouts.Value `tfsdk:"timeouts"`
}

type clusterCiliumModel struct {
	EnvoyDaemonset types.Bool `tfsdk:"envoy_daemonset"`
	HubbleRelay    types.Bool `tfsdk:"hubble_relay"`
}

type clusterKubeOptionsModel struct {
	FeatureGates         types.Set    `tfsdk:"feature_gates"`
	AdmissionControllers types.Set    `tfsdk:"admission_controllers"`
	AuditLogs            types.Object `tfsdk:"audit_logs"`
	OIDC                 types.Object `tfsdk:"oidc"`
}

type clusterAuditLogsModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type clusterOIDCModel struct {
	Enabled       types.Bool         `tfsdk:"enabled"`
	ProviderName  types.String       `tfsdk:"provider_name"`
	IssuerURL     types.String       `tfsdk:"issuer_url"`
	ClientID      types.String       `tfsdk:"client_id"`
	UsernameClaim types.String       `tfsdk:"username_claim"`
	GroupsClaim   types.String       `tfsdk:"groups_claim"`
	CACerts       trimmedStringValue `tfsdk:"ca_certs"`
}

// compareKubeVersions ignores the patch version when one of the versions has none.
func compareKubeVersions(a, b string) (int, error) {
	parts := []func(string) (int, error){kubeVersionToMajor, kubeVersionToMinor}
	if !kubeVersionIsMinor(a) && !kubeVersionIsMinor(b) {
		parts = append(parts, kubeVersionToPatch)
	}
	for _, part := range parts {
		partA, err := part(a)
		if err != nil {
			return 0, err
		}
		partB, err := part(b)
		if err != nil {
			return 0, err
		}
		order := cmp.Compare(partA, partB)
		if order != 0 {
			return order, nil
		}
	}

	return 0, nil
}

func expandClusterCreateOpts(ctx context.Context, plan clusterModel) (*mksclient.ClusterCreateStruct, diag.Diagnostics) {
	opts := &mksclient.ClusterCreateStruct{
		Name:                          plan.Name.ValueString(),
		Pool:                          plan.Pool.ValueString(),
		KubeVersion:                   plan.KubeVersion.ValueString(),
		NetworkType:                   workersTypes[plan.WorkersType.ValueString()],
		NetworkId:                     plan.NetworkID.ValueString(),
		SubnetId:                      plan.SubnetID.ValueString(),
		CloudSubnetCidr:               plan.CloudSubnetCIDR.ValueString(),
		PrivateKubeApi:                plan.PrivateKubeAPI.ValueBool(),
		MaintenanceWindowStart:        plan.MaintenanceWindowStart.ValueString(),
		EnableAutorepair:              knownBoolPointer(plan.EnableAutorepair),
		EnablePatchVersionAutoUpgrade: knownBoolPointer(plan.EnablePatchVersionAutoUpgrade),
	}
	if !plan.ClusterType.IsNull() && !plan.ClusterType.IsUnknown() {
		clusterType := mksclient.ClusterCreateStructClusterType(plan.ClusterType.ValueString())
		opts.ClusterType = &clusterType
	}

	var diags diag.Diagnostics
	if !plan.CNIType.IsUnknown() {
		opts.CniType = plan.CNIType.ValueStringPointer()
	}
	opts.CniCiliumSettings, diags = expandCilium(ctx, plan.CNICiliumSettings)
	kubeOptions, kubeOptionsDiags := expandKubeOptions(ctx, plan.KubernetesOptions)
	diags.Append(kubeOptionsDiags...)
	opts.KubernetesOptions = kubeOptions

	return opts, diags
}

func expandClusterPatch(ctx context.Context, plan, state clusterModel) (*mksclient.ClusterUpdateStruct, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	patch := &mksclient.ClusterUpdateStruct{}
	changed := false

	if !plan.MaintenanceWindowStart.IsUnknown() && !plan.MaintenanceWindowStart.Equal(state.MaintenanceWindowStart) {
		patch.MaintenanceWindowStart = plan.MaintenanceWindowStart.ValueStringPointer()
		changed = true
	}
	if !plan.EnableAutorepair.IsUnknown() && !plan.EnableAutorepair.Equal(state.EnableAutorepair) {
		patch.EnableAutorepair = plan.EnableAutorepair.ValueBoolPointer()
		changed = true
	}
	if !plan.EnablePatchVersionAutoUpgrade.IsUnknown() && !plan.EnablePatchVersionAutoUpgrade.Equal(state.EnablePatchVersionAutoUpgrade) {
		patch.EnablePatchVersionAutoUpgrade = plan.EnablePatchVersionAutoUpgrade.ValueBoolPointer()
		changed = true
	}
	if !plan.CNICiliumSettings.Equal(state.CNICiliumSettings) {
		cilium, ciliumDiags := expandCilium(ctx, plan.CNICiliumSettings)
		diags.Append(ciliumDiags...)
		patch.CniCiliumSettings = cilium
		changed = changed || cilium != nil
	}
	if !plan.KubernetesOptions.Equal(state.KubernetesOptions) {
		kubeOptions, kubeOptionsDiags := expandKubeOptions(ctx, plan.KubernetesOptions)
		diags.Append(kubeOptionsDiags...)
		patch.KubernetesOptions = kubeOptions
		changed = changed || kubeOptions != nil
	}

	return patch, changed, diags
}

func expandCilium(ctx context.Context, obj types.Object) (*mksclient.CNICiliumSettings, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	var m clusterCiliumModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})

	return &mksclient.CNICiliumSettings{
		EnvoyDaemonset: knownBoolPointer(m.EnvoyDaemonset),
		HubbleRelay:    knownBoolPointer(m.HubbleRelay),
	}, diags
}

func expandKubeOptions(ctx context.Context, obj types.Object) (*mksclient.KubernetesOptions, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	var m clusterKubeOptionsModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	opts := &mksclient.KubernetesOptions{
		FeatureGates:         []string{},
		AdmissionControllers: []string{},
	}
	if !m.FeatureGates.IsNull() && !m.FeatureGates.IsUnknown() {
		diags.Append(m.FeatureGates.ElementsAs(ctx, &opts.FeatureGates, false)...)
	}
	if !m.AdmissionControllers.IsNull() && !m.AdmissionControllers.IsUnknown() {
		diags.Append(m.AdmissionControllers.ElementsAs(ctx, &opts.AdmissionControllers, false)...)
	}
	if !m.AuditLogs.IsNull() && !m.AuditLogs.IsUnknown() {
		var audit clusterAuditLogsModel
		diags.Append(m.AuditLogs.As(ctx, &audit, basetypes.ObjectAsOptions{})...)
		opts.AuditLogs = mksclient.AuditLogs{Enabled: audit.Enabled.ValueBool()}
	}
	if !m.OIDC.IsNull() && !m.OIDC.IsUnknown() {
		var oidc clusterOIDCModel
		diags.Append(m.OIDC.As(ctx, &oidc, basetypes.ObjectAsOptions{})...)
		opts.Oidc = mksclient.OIDC{
			Enabled:       oidc.Enabled.ValueBool(),
			ProviderName:  oidc.ProviderName.ValueString(),
			IssuerUrl:     oidc.IssuerURL.ValueString(),
			ClientId:      oidc.ClientID.ValueString(),
			UsernameClaim: oidc.UsernameClaim.ValueString(),
			GroupsClaim:   oidc.GroupsClaim.ValueString(),
			CaCerts:       oidc.CACerts.ValueString(),
		}
		if !opts.Oidc.Enabled {
			// The API rejects any parameter of a disabled OIDC and clears them itself.
			opts.Oidc = mksclient.OIDC{}
		}
	}

	return opts, diags
}

func (m *clusterModel) fromAPI(ctx context.Context, c *mksclient.ClusterDetailed, prior clusterModel, applied bool) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(c.Id)
	m.Name = types.StringValue(c.Name)
	if strings.EqualFold(prior.Name.ValueString(), c.Name) {
		// The API lower-cases the name, see nameCase.
		m.Name = prior.Name
	}
	m.Pool = types.StringValue(c.Pool)
	if c.ProjectId != "" && !sameProject(prior.ProjectID.ValueString(), c.ProjectId) {
		// The API returns the project dashed: the state takes the form keystoneProjectID keeps.
		m.ProjectID = types.StringValue(keystoneProjectID(c.ProjectId))
	}
	apiMinor, _ := kubeVersionTrimToMinor(c.KubeVersion)
	// A state version without the patch stays: it stands for any patch of its minor.
	if !applied && (apiMinor == "" || prior.KubeVersion.ValueString() != apiMinor) {
		m.KubeVersion = types.StringValue(c.KubeVersion)
	}
	m.ClusterType = types.StringValue(string(c.ClusterType))
	m.WorkersType = types.StringValue(string(c.NetworkType))
	for workersType, networkType := range workersTypes {
		if networkType == string(c.NetworkType) {
			m.WorkersType = types.StringValue(workersType)
		}
	}
	m.NetworkID = types.StringValue(c.NetworkId)
	m.SubnetID = types.StringValue(c.SubnetId)
	m.CloudSubnetCIDR = types.StringPointerValue(c.CloudSubnetCidr)
	m.PrivateKubeAPI = types.BoolValue(c.PrivateKubeApi)
	m.MaintenanceWindowStart = types.StringValue(c.MaintenanceWindowStart)
	m.MaintenanceWindowEnd = types.StringValue(c.MaintenanceWindowEnd)
	m.EnableAutorepair = types.BoolValue(c.EnableAutorepair)
	m.EnablePatchVersionAutoUpgrade = types.BoolValue(c.EnablePatchVersionAutoUpgrade)
	m.CNIType = types.StringValue(string(c.CniType))
	m.Status = types.StringValue(string(c.Status))
	m.KubeAPIIP = types.StringValue(c.KubeApiIp)

	m.CNICiliumSettings = types.ObjectNull(ciliumAttrTypes)
	if c.CniCiliumSettings != nil {
		var d diag.Diagnostics
		m.CNICiliumSettings, d = types.ObjectValueFrom(ctx, ciliumAttrTypes, clusterCiliumModel{
			EnvoyDaemonset: types.BoolPointerValue(c.CniCiliumSettings.EnvoyDaemonset),
			HubbleRelay:    types.BoolPointerValue(c.CniCiliumSettings.HubbleRelay),
		})
		diags.Append(d...)
	}

	kubeOptions, d := flattenKubeOptions(ctx, c.KubernetesOptions)
	diags.Append(d...)
	m.KubernetesOptions = kubeOptions

	return diags
}

// checkRemovedKubeOptions fails for the planned options that the upgraded cluster does not have.
func checkRemovedKubeOptions(ctx context.Context, planned types.Object, got *mksclient.ClusterDetailed) diag.Diagnostics {
	want, diags := expandKubeOptions(ctx, planned)
	if want == nil {
		return diags
	}

	for _, option := range []struct {
		name          string
		planned, kept []string
	}{
		{name: "feature_gates", planned: want.FeatureGates, kept: got.KubernetesOptions.FeatureGates},
		{name: "admission_controllers", planned: want.AdmissionControllers, kept: got.KubernetesOptions.AdmissionControllers},
	} {
		removed := slices.DeleteFunc(option.planned, func(name string) bool {
			return slices.Contains(option.kept, name)
		})
		if len(removed) > 0 {
			diags.AddAttributeError(path.Root("kubernetes_options").AtName(option.name),
				"Kubernetes options are missing after the upgrade",
				fmt.Sprintf("The cluster runs Kubernetes %s without %s. If the new version does not have them, remove them from %s.",
					got.KubeVersion, strings.Join(removed, ", "), option.name))
		}
	}

	return diags
}

func flattenKubeOptions(ctx context.Context, o mksclient.KubernetesOptions) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	stringSet := func(values []string) types.Set {
		// A framework set rejects duplicates.
		values = slices.Compact(slices.Sorted(slices.Values(values)))
		set, d := types.SetValueFrom(ctx, types.StringType, append([]string{}, values...))
		diags.Append(d...)

		return set
	}
	audit, d := types.ObjectValueFrom(ctx, auditLogsAttrTypes, clusterAuditLogsModel{
		Enabled: types.BoolValue(o.AuditLogs.Enabled),
	})
	diags.Append(d...)
	oidcObject, d := types.ObjectValueFrom(ctx, oidcAttrTypes, clusterOIDCModel{
		Enabled:       types.BoolValue(o.Oidc.Enabled),
		ProviderName:  types.StringValue(o.Oidc.ProviderName),
		IssuerURL:     types.StringValue(o.Oidc.IssuerUrl),
		ClientID:      types.StringValue(o.Oidc.ClientId),
		UsernameClaim: types.StringValue(o.Oidc.UsernameClaim),
		GroupsClaim:   types.StringValue(o.Oidc.GroupsClaim),
		CACerts:       trimmedString(o.Oidc.CaCerts),
	})
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, kubeOptionsAttrTypes, clusterKubeOptionsModel{
		FeatureGates:         stringSet(o.FeatureGates),
		AdmissionControllers: stringSet(o.AdmissionControllers),
		AuditLogs:            audit,
		OIDC:                 oidcObject,
	})
	diags.Append(d...)

	return obj, diags
}

// knownBoolPointer returns nil for an unknown bool, which ValueBoolPointer
// turns into false.
func knownBoolPointer(v types.Bool) *bool {
	if v.IsUnknown() {
		return nil
	}

	return v.ValueBoolPointer()
}

// plannedState saves the plan with its unknown values as null, so Terraform
// keeps a created cluster it could not read instead of losing it.
func plannedState(plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var diags diag.Diagnostics

	raw, err := tftypes.Transform(plan.Raw, func(_ *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if !v.IsKnown() {
			return tftypes.NewValue(v.Type(), nil), nil
		}

		return v, nil
	})
	if err != nil {
		diags.AddError("Error saving the planned state", err.Error())

		return diags
	}
	state.Raw = raw

	return diags
}
