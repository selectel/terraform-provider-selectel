package mks

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterOIDCUnknownWhileDisabled(t *testing.T) {
	fake := newFakeAPI(t)
	r := fake.resource()

	plan := clusterPlan()
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": enabledOIDCObject(),
	})
	state, diags := createCluster(t, r, plan)
	require.False(t, diags.HasError(), "%v", diags)
	prior := readClusterModel(t, state)

	newPlan := prior
	newPlan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"oidc": types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
			"enabled":        types.BoolValue(false),
			"provider_name":  types.StringNull(),
			"issuer_url":     types.StringUnknown(),
			"client_id":      types.StringNull(),
			"username_claim": types.StringNull(),
			"groups_claim":   types.StringNull(),
			"ca_certs":       trimmedString(""),
		}),
	})
	_, diags = updateCluster(t, r, prior, newPlan)
	require.False(t, diags.HasError(), "%v", diags)
	oidc := oidcOf(fake.lastBody(routePatchCluster))
	assert.Equal(t, map[string]any{
		"enabled": false, "provider_name": "", "issuer_url": "", "client_id": "", "username_claim": "", "groups_claim": "",
	}, oidc)

	modReq := planmodifier.StringRequest{
		Config:      tfsdk.Config(clusterStateOf(t, r, newPlan)),
		State:       clusterStateOf(t, r, prior),
		ConfigValue: types.StringNull(),
		StateValue:  types.StringValue("https://issuer.example.com"),
		Path:        path.Root("kubernetes_options").AtName("oidc").AtName("issuer_url"),
	}
	resp := planmodifier.StringResponse{PlanValue: types.StringNull()}
	oidcPlanModifier{}.PlanModifyString(context.Background(), modReq, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.True(t, !resp.PlanValue.IsNull() && resp.PlanValue.ValueString() == "")
}
func TestClusterCiliumSettingsNeedCilium(t *testing.T) {
	for _, tt := range []struct {
		name    string
		cniType types.String
	}{
		{name: "cni_type null"},
		{name: "cni_type CALICO", cniType: types.StringValue("CALICO")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t)
			r := fake.resource()
			plan := clusterPlan()
			plan.CNIType = tt.cniType
			plan.CNICiliumSettings = types.ObjectValueMust(ciliumAttrTypes, map[string]attr.Value{
				"envoy_daemonset": types.BoolValue(true),
				"hubble_relay":    types.BoolValue(false),
			})
			config := tfsdk.Config(clusterStateOf(t, r, plan))
			resp := resource.ValidateConfigResponse{}
			r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: config}, &resp)

			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, "Cilium settings need the Cilium CNI", resp.Diagnostics.Errors()[0].Summary())
		})
	}
}
func TestClusterModifyPlanKubeVersion(t *testing.T) {
	for _, tt := range []struct {
		name  string
		state string
		plan  string
	}{
		{name: "lower same minor", state: "1.31.8", plan: "1.31.5"},
		{name: "lower minor", state: "1.32.8", plan: "1.31.5"},
		{name: "lower major", state: "2.31.1", plan: "1.31.5"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := &clusterResource{}
			state := clusterState()
			state.KubeVersion = types.StringValue(tt.state)
			plan := clusterState()
			plan.KubeVersion = types.StringValue(tt.plan)
			req := resource.ModifyPlanRequest{
				Config: tfsdk.Config(clusterStateOf(t, r, plan)), Plan: tfsdk.Plan(clusterStateOf(t, r, plan)),
				State: clusterStateOf(t, r, state),
			}
			resp := resource.ModifyPlanResponse{Plan: req.Plan}
			r.ModifyPlan(context.Background(), req, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var got clusterModel
			require.False(t, resp.Plan.Get(context.Background(), &got).HasError())
			assert.Equal(t, types.StringValue(tt.state), got.KubeVersion)
		})
	}
}
func TestClusterX509RemovalRejectedAtPlan(t *testing.T) {
	r := &clusterResource{}
	state := clusterState()
	state.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringValue("Y2VydA=="),
	})
	plan := state
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringNull(),
	})
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, plan)), Plan: tfsdk.Plan(clusterStateOf(t, r, plan)),
		State: clusterStateOf(t, r, state),
	}
	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(context.Background(), req, &resp)
	require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, "X509 CA certificates cannot be removed", resp.Diagnostics.Errors()[0].Summary())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "The API keeps the X509 CA certificates")
}
func TestClusterModifyPlanKeepsStatusWhenPlanEqualsState(t *testing.T) {
	r := &clusterResource{}
	state := clusterState()
	plan := state
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, plan)), Plan: tfsdk.Plan(clusterStateOf(t, r, plan)),
		State: clusterStateOf(t, r, state),
	}
	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(context.Background(), req, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got clusterModel
	require.False(t, resp.Plan.Get(context.Background(), &got).HasError())
	assert.Equal(t, types.StringValue("ACTIVE"), got.Status)
	assert.Equal(t, types.StringValue("09:00:00"), got.MaintenanceWindowEnd)
}
func TestClusterBasicClusterAutoUpgradeRejectedAtPlan(t *testing.T) {
	r := &clusterResource{}
	config := clusterPlan()
	config.ClusterType = types.StringValue("BASIC")
	config.EnablePatchVersionAutoUpgrade = types.BoolValue(true)
	plan := tfsdk.Config(clusterStateOf(t, r, config))
	resp := resource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: plan}, &resp)
	require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, "Patch version auto-upgrade is not available for BASIC clusters", resp.Diagnostics.Errors()[0].Summary())
}
func TestClusterValidateConfigRules(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(m *clusterModel)
		want   string
	}{
		{
			name: "cloud cluster rejects cloud_subnet_cidr",
			mutate: func(m *clusterModel) {
				m.WorkersType = types.StringValue("CLOUD")
				m.CloudSubnetCIDR = types.StringValue("10.20.0.0/24")
			},
			want: "Cloud subnet CIDR is not available for cloud clusters",
		},
		{
			name: "dedicated cluster rejects enable_autorepair true",
			mutate: func(m *clusterModel) {
				m.WorkersType = types.StringValue("DEDICATED")
				m.EnableAutorepair = types.BoolValue(true)
			},
			want: "Auto-repair is not available for dedicated clusters",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := &clusterResource{}
			config := clusterPlan()
			tt.mutate(&config)
			resp := resource.ValidateConfigResponse{}
			r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{
				Config: tfsdk.Config(clusterStateOf(t, r, config)),
			}, &resp)
			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tt.want, resp.Diagnostics.Errors()[0].Summary())
		})
	}
}
func TestClusterValidateConfigAcceptsDedicatedCIDR(t *testing.T) {
	r := &clusterResource{}
	config := clusterPlan()
	config.WorkersType = types.StringValue("DEDICATED")
	config.EnableAutorepair = types.BoolValue(false)
	config.CloudSubnetCIDR = types.StringValue("10.20.0.0/24")
	resp := resource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, config)),
	}, &resp)
	assert.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
}
func TestClusterX509UnknownPlanSkipsTheCheck(t *testing.T) {
	r := &clusterResource{}
	state := clusterState()
	state.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringValue("Y2VydA=="),
	})
	plan := state
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringUnknown(),
	})
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, plan)), Plan: tfsdk.Plan(clusterStateOf(t, r, plan)),
		State: clusterStateOf(t, r, state),
	}
	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(context.Background(), req, &resp)
	assert.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
}
func TestClusterX509UnknownObjectSkipsTheCheck(t *testing.T) {
	r := &clusterResource{}
	state := clusterState()
	state.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
		"x509_ca_certificates": types.StringValue("Y2VydA=="),
	})
	unknownPlan := state
	unknownPlan.KubernetesOptions = types.ObjectUnknown(kubeOptionsAttrTypes)
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, unknownPlan)), Plan: tfsdk.Plan(clusterStateOf(t, r, unknownPlan)),
		State: clusterStateOf(t, r, state),
	}
	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(context.Background(), req, &resp)
	assert.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
}
func TestClusterAuditSecretModifierRules(t *testing.T) {
	r := &clusterResource{}
	auditLogs := func(enabled types.Bool, secret types.String) types.Object {
		return types.ObjectValueMust(auditLogsAttrTypes, map[string]attr.Value{
			"enabled": enabled, "secret_name": secret,
		})
	}
	reqOf := func(t *testing.T, planned, configAudit types.Object) planmodifier.StringRequest {
		t.Helper()
		plannedModel := clusterPlan()
		plannedModel.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"audit_logs": planned})
		configModel := plannedModel
		configModel.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"audit_logs": configAudit})
		st := clusterStateOf(t, r, plannedModel)
		req := planmodifier.StringRequest{
			Config: tfsdk.Config(clusterStateOf(t, r, configModel)),
			Plan:   tfsdk.Plan(st),
			State:  st,
			Path:   path.Root("kubernetes_options").AtName("audit_logs").AtName("secret_name"),
		}

		return req
	}
	for _, tt := range []struct {
		name    string
		planned types.Object
		config  types.Object
		configV types.String
		stateV  types.String
		want    types.String
		wantErr string
	}{
		{
			name:    "a non-empty name with audit logs planned disabled fails the plan",
			planned: auditLogs(types.BoolValue(false), types.StringNull()),
			config:  auditLogs(types.BoolValue(true), types.StringValue("siem")),
			configV: types.StringValue("siem"),
			wantErr: "Audit logs secret name set while audit logs are disabled",
		},
		{
			name:    "planned disabled with a null config name plans empty",
			planned: auditLogs(types.BoolValue(false), types.StringNull()),
			config:  auditLogs(types.BoolValue(false), types.StringNull()),
			configV: types.StringNull(),
			want:    types.StringValue(""),
		},
		{
			name:    "a null name with audit logs planned enabled plans the API default",
			planned: auditLogs(types.BoolValue(true), types.StringNull()),
			config:  auditLogs(types.BoolValue(true), types.StringNull()),
			configV: types.StringNull(),
			want:    types.StringValue("mks-audit-logs"),
		},
		{
			name:    "planned unknown enabled with a null config name plans unknown",
			planned: auditLogs(types.BoolUnknown(), types.StringNull()),
			config:  auditLogs(types.BoolNull(), types.StringNull()),
			configV: types.StringNull(),
			want:    types.StringUnknown(),
		},
		{
			name:    "omitted enabled on an enabled state keeps the state name",
			planned: auditLogs(types.BoolValue(true), types.StringValue("siem")),
			config:  auditLogs(types.BoolNull(), types.StringNull()),
			configV: types.StringNull(),
			stateV:  types.StringValue("siem"),
			want:    types.StringValue("siem"),
		},
		{
			name:    "omitted enabled on a disabled state plans an empty name",
			planned: auditLogs(types.BoolValue(false), types.StringNull()),
			config:  auditLogs(types.BoolNull(), types.StringNull()),
			configV: types.StringNull(),
			stateV:  types.StringValue("siem"),
			want:    types.StringValue(""),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := reqOf(t, tt.planned, tt.config)
			req.ConfigValue = tt.configV
			req.StateValue = types.StringNull()
			if !tt.stateV.IsNull() {
				req.StateValue = tt.stateV
			}
			resp := planmodifier.StringResponse{PlanValue: req.ConfigValue}
			auditSecretPlanModifier{}.PlanModifyString(context.Background(), req, &resp)
			if tt.wantErr != "" {
				require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				assert.Equal(t, tt.wantErr, resp.Diagnostics.Errors()[0].Summary())

				return
			}
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tt.want, resp.PlanValue)
		})
	}
}
func TestClusterOIDCModifierRules(t *testing.T) {
	r := &clusterResource{}
	oidc := func(enabled bool) types.Object {
		return types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
			"enabled":        types.BoolValue(enabled),
			"provider_name":  types.StringValue("keycloak"),
			"issuer_url":     types.StringValue("https://issuer.example.com"),
			"client_id":      types.StringValue("kubernetes"),
			"username_claim": types.StringNull(),
			"groups_claim":   types.StringNull(),
			"ca_certs":       trimmedStringValue{StringValue: types.StringNull()},
		})
	}
	oidcWithEnabledAndClaim := func(enabled types.Bool, claim types.String) types.Object {
		return types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
			"enabled":        enabled,
			"provider_name":  types.StringNull(),
			"issuer_url":     types.StringNull(),
			"client_id":      types.StringNull(),
			"username_claim": claim,
			"groups_claim":   types.StringNull(),
			"ca_certs":       trimmedStringValue{StringValue: types.StringNull()},
		})
	}
	t.Run("OIDC parameter set while OIDC is disabled", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidc(false)})
		modReq := planmodifier.StringRequest{
			Config:      tfsdk.Config(clusterStateOf(t, r, config)),
			ConfigValue: types.StringValue("keycloak"),
			StateValue:  types.StringNull(),
			Path:        path.Root("kubernetes_options").AtName("oidc").AtName("provider_name"),
		}
		resp := planmodifier.StringResponse{PlanValue: types.StringValue("keycloak")}
		oidcPlanModifier{}.PlanModifyString(context.Background(), modReq, &resp)
		require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, "OIDC parameter set while OIDC is disabled", resp.Diagnostics.Errors()[0].Summary())
		assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "oidc parameters cannot be configured when it is disabled")
	})
	for _, name := range []string{"username_claim", "groups_claim"} {
		t.Run(name+" cannot be empty while enabled", func(t *testing.T) {
			enabled := oidc(true)
			enabledObj := clusterPlan()
			enabledObj.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": enabled})
			modReq := planmodifier.StringRequest{
				Config:      tfsdk.Config(clusterStateOf(t, r, enabledObj)),
				ConfigValue: types.StringValue(""),
				StateValue:  types.StringNull(),
				Path:        path.Root("kubernetes_options").AtName("oidc").AtName(name),
			}
			resp := planmodifier.StringResponse{PlanValue: types.StringValue("")}
			oidcPlanModifier{enabledDefault: "sub"}.PlanModifyString(context.Background(), modReq, &resp)
			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, "Empty OIDC claim while OIDC is enabled", resp.Diagnostics.Errors()[0].Summary())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), name+" cannot be empty while OIDC is enabled: omit it to get")
		})
	}
	t.Run("null enabled with an enabled state keeps the plan", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
			"oidc": oidcWithEnabledAndClaim(types.BoolNull(), types.StringValue("https://issuer.example.com")),
		})
		state := clusterState()
		state.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidc(true)})
		modReq := planmodifier.StringRequest{
			Config:      tfsdk.Config(clusterStateOf(t, r, config)),
			State:       clusterStateOf(t, r, state),
			ConfigValue: types.StringValue("https://issuer.example.com"),
			StateValue:  types.StringValue("https://issuer.example.com"),
			Path:        path.Root("kubernetes_options").AtName("oidc").AtName("issuer_url"),
		}
		resp := planmodifier.StringResponse{PlanValue: modReq.ConfigValue}
		oidcPlanModifier{}.PlanModifyString(context.Background(), modReq, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, types.StringValue("https://issuer.example.com"), resp.PlanValue)
	})
	t.Run("null enabled on create plans an empty parameter", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
			"oidc": oidcWithEnabledAndClaim(types.BoolNull(), types.StringNull()),
		})
		modReq := planmodifier.StringRequest{
			Config:      tfsdk.Config(clusterStateOf(t, r, config)),
			ConfigValue: types.StringNull(),
			StateValue:  types.StringNull(),
			Path:        path.Root("kubernetes_options").AtName("oidc").AtName("issuer_url"),
		}
		resp := planmodifier.StringResponse{PlanValue: types.StringUnknown()}
		oidcPlanModifier{}.PlanModifyString(context.Background(), modReq, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, types.StringValue(""), resp.PlanValue)
	})
	t.Run("null enabled with an enabled state and a null claim plans the default", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
			"oidc": oidcWithEnabledAndClaim(types.BoolNull(), types.StringNull()),
		})
		state := clusterState()
		state.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidc(true)})
		modReq := planmodifier.StringRequest{
			Config:      tfsdk.Config(clusterStateOf(t, r, config)),
			State:       clusterStateOf(t, r, state),
			ConfigValue: types.StringNull(),
			StateValue:  types.StringNull(),
			Path:        path.Root("kubernetes_options").AtName("oidc").AtName("username_claim"),
		}
		resp := planmodifier.StringResponse{PlanValue: types.StringUnknown()}
		oidcPlanModifier{enabledDefault: "sub"}.PlanModifyString(context.Background(), modReq, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, types.StringValue("sub"), resp.PlanValue)
	})
}
func TestClusterOIDCKeepsStateWhileEnabled(t *testing.T) {
	r := &clusterResource{}
	plan := clusterPlan()
	enabledOIDC := types.ObjectValueMust(oidcAttrTypes, map[string]attr.Value{
		"enabled":        types.BoolValue(true),
		"provider_name":  types.StringNull(),
		"issuer_url":     types.StringNull(),
		"client_id":      types.StringNull(),
		"username_claim": types.StringNull(),
		"groups_claim":   types.StringNull(),
		"ca_certs":       trimmedStringValue{StringValue: types.StringNull()},
	})
	plan.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": enabledOIDC})

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	kubeOptions := schemaResp.Schema.Attributes["kubernetes_options"].(schema.SingleNestedAttribute)
	oidc := kubeOptions.Attributes["oidc"].(schema.SingleNestedAttribute)
	attr := oidc.Attributes["provider_name"].(schema.StringAttribute)
	require.NotEmpty(t, attr.PlanModifiers)

	modReq := planmodifier.StringRequest{
		Config:      tfsdk.Config(clusterStateOf(t, r, plan)),
		State:       clusterStateOf(t, r, plan),
		PlanValue:   types.StringUnknown(),
		StateValue:  types.StringValue("keycloak"),
		ConfigValue: types.StringNull(),
		Path:        path.Root("kubernetes_options").AtName("oidc").AtName("provider_name"),
	}
	resp := planmodifier.StringResponse{PlanValue: modReq.PlanValue}
	for _, m := range attr.PlanModifiers {
		m.PlanModifyString(context.Background(), modReq, &resp)
		modReq.PlanValue = resp.PlanValue
	}
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, types.StringValue("keycloak"), resp.PlanValue)
}
func TestClusterModifyPlanSetsUnknowns(t *testing.T) {
	r := &clusterResource{}
	changePlan := clusterState()
	changePlan.KubeVersion = types.StringValue(testKubeVersionUp)
	changePlan.MaintenanceWindowStart = types.StringValue("07:00:00")
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, changePlan)),
		Plan:   tfsdk.Plan(clusterStateOf(t, r, changePlan)), State: clusterStateOf(t, r, clusterState()),
	}
	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(context.Background(), req, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got clusterModel
	require.False(t, resp.Plan.Get(context.Background(), &got).HasError())
	assert.True(t, got.Status.IsUnknown())
	assert.True(t, got.MaintenanceWindowEnd.IsUnknown())
}
func TestClusterNameCaseModifier(t *testing.T) {
	for _, tt := range []struct {
		name     string
		plan     types.String
		expected types.String
	}{
		{name: "case-only change keeps the state name", plan: types.StringValue("Test-Cluster"), expected: types.StringValue("test-cluster")},
		{name: "a different name is not rewritten", plan: types.StringValue("other-cluster"), expected: types.StringValue("other-cluster")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := planmodifier.StringResponse{PlanValue: tt.plan}
			nameCase{}.PlanModifyString(context.Background(), planmodifier.StringRequest{
				StateValue: types.StringValue("test-cluster"), PlanValue: tt.plan,
			}, &resp)
			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}
func TestClusterSameProjectIDModifier(t *testing.T) {
	for _, tt := range []struct {
		name     string
		plan     types.String
		expected types.String
	}{
		{name: "dashed form keeps the state project", plan: types.StringValue("a07abc12-3105-46f1-b929-1ab301a7d75b"), expected: types.StringValue("a07abc12310546f1b9291ab301a7d75b")},
		{name: "another project is not rewritten", plan: types.StringValue("c07abc12310546f1b9291ab301a7d75b"), expected: types.StringValue("c07abc12310546f1b9291ab301a7d75b")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := planmodifier.StringResponse{PlanValue: tt.plan}
			sameProjectID{}.PlanModifyString(context.Background(), planmodifier.StringRequest{
				StateValue: types.StringValue("a07abc12310546f1b9291ab301a7d75b"), PlanValue: tt.plan,
			}, &resp)
			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}

	t.Run("null state keeps the plan", func(t *testing.T) {
		resp := planmodifier.StringResponse{PlanValue: types.StringValue("x")}
		sameProjectID{}.PlanModifyString(context.Background(), planmodifier.StringRequest{
			StateValue: types.StringNull(),
		}, &resp)
		assert.Equal(t, types.StringValue("x"), resp.PlanValue)
	})
}
func TestClusterTrimmedStringKeepModifier(t *testing.T) {
	for _, tt := range []struct {
		name     string
		config   types.String
		expected types.String
	}{
		{name: "surrounding whitespace keeps the state value", config: types.StringValue("cert-1\n"), expected: types.StringValue("cert-1")},
		{name: "a different value is not rewritten", config: types.StringValue("cert-2"), expected: types.StringValue("cert-2")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := planmodifier.StringResponse{PlanValue: tt.config}
			trimmedStringKeep{}.PlanModifyString(context.Background(), planmodifier.StringRequest{
				StateValue: types.StringValue("cert-1"), ConfigValue: tt.config, PlanValue: tt.config,
			}, &resp)
			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}
