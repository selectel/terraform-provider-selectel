package mks

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validateConfig(t *testing.T, r *clusterResource, config clusterModel) resource.ValidateConfigResponse {
	t.Helper()
	resp := resource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, config)),
	}, &resp)

	return resp
}

func TestClusterValidateConfigRules(t *testing.T) {
	r := &clusterResource{}
	for _, tt := range []struct {
		name   string
		mutate func(m *clusterModel)
		want   string
	}{
		{name: "cilium settings need the Cilium CNI", mutate: func(m *clusterModel) {
			m.CNICiliumSettings = ciliumObject(types.BoolValue(true), types.BoolValue(false))
		}, want: "Cilium settings need the Cilium CNI"},
		{name: "BASIC rejects patch version auto-upgrade", mutate: func(m *clusterModel) {
			m.ClusterType, m.EnablePatchVersionAutoUpgrade = types.StringValue("BASIC"), types.BoolValue(true)
		}, want: "Patch version auto-upgrade is not available for BASIC clusters"},
		{name: "cloud cluster rejects cloud_subnet_cidr", mutate: func(m *clusterModel) {
			m.WorkersType, m.CloudSubnetCIDR = types.StringValue("CLOUD"), types.StringValue("10.20.0.0/24")
		}, want: "Cloud subnet CIDR is not available for cloud clusters"},
		{name: "dedicated cluster rejects enable_autorepair true", mutate: func(m *clusterModel) {
			m.WorkersType, m.EnableAutorepair = types.StringValue("DEDICATED"), types.BoolValue(true)
		}, want: "Auto-repair is not available for dedicated clusters"},
		{name: "dedicated cluster accepts cloud_subnet_cidr and no autorepair", mutate: func(m *clusterModel) {
			m.WorkersType, m.EnableAutorepair = types.StringValue("DEDICATED"), types.BoolValue(false)
			m.CloudSubnetCIDR = types.StringValue("10.20.0.0/24")
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			config := clusterPlan()
			tt.mutate(&config)
			resp := validateConfig(t, r, config)
			if tt.want == "" {
				assert.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

				return
			}
			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tt.want, resp.Diagnostics.Errors()[0].Summary())
		})
	}
}

func modifyPlan(t *testing.T, r *clusterResource, state, plan clusterModel) resource.ModifyPlanResponse {
	t.Helper()
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config(clusterStateOf(t, r, plan)), Plan: tfsdk.Plan(clusterStateOf(t, r, plan)),
		State: clusterStateOf(t, r, state),
	}
	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(context.Background(), req, &resp)

	return resp
}

func TestClusterModifyPlanX509Rules(t *testing.T) {
	r := &clusterResource{}
	withX509 := func(m *clusterModel) {
		m.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"x509_ca_certificates": types.StringValue("Y2VydA==")})
	}
	for _, tt := range []struct {
		name    string
		plan    func(m *clusterModel)
		wantErr string
	}{
		{name: "null x509 over a set one is a removal", plan: func(m *clusterModel) {
			m.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"x509_ca_certificates": types.StringNull()})
		}, wantErr: "X509 CA certificates cannot be removed"},
		{name: "unknown x509 skips the check", plan: func(m *clusterModel) {
			m.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"x509_ca_certificates": types.StringUnknown()})
		}},
		{name: "unknown kubernetes_options skips the check", plan: func(m *clusterModel) {
			m.KubernetesOptions = types.ObjectUnknown(kubeOptionsAttrTypes)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := clusterState()
			withX509(&state)
			plan := state
			tt.plan(&plan)
			resp := modifyPlan(t, r, state, plan)
			if tt.wantErr == "" {
				assert.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

				return
			}
			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tt.wantErr, resp.Diagnostics.Errors()[0].Summary())
		})
	}
}

func TestClusterModifyPlanState(t *testing.T) {
	r := &clusterResource{}
	for _, tt := range []struct {
		name   string
		mutate func(m *clusterModel)
		check  func(t *testing.T, plan clusterModel)
	}{
		{name: "a lower kube_version keeps the state version", mutate: func(m *clusterModel) {
			m.KubeVersion = types.StringValue("1.31.5")
		}, check: func(t *testing.T, plan clusterModel) {
			assert.Equal(t, types.StringValue("1.31.8"), plan.KubeVersion)
		}},
		{name: "plan equal to state keeps the known values", mutate: func(_ *clusterModel) {}, check: func(t *testing.T, plan clusterModel) {
			assert.Equal(t, types.StringValue("ACTIVE"), plan.Status)
			assert.Equal(t, types.StringValue("09:00:00"), plan.MaintenanceWindowEnd)
		}},
		{name: "a change sets the computed values unknown", mutate: func(m *clusterModel) {
			m.KubeVersion = types.StringValue(testKubeVersionUp)
			m.MaintenanceWindowStart = types.StringValue("07:00:00")
		}, check: func(t *testing.T, plan clusterModel) {
			assert.True(t, plan.Status.IsUnknown())
			assert.True(t, plan.MaintenanceWindowEnd.IsUnknown())
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := modifyPlan(t, r, clusterState(), func() clusterModel {
				plan := clusterState()
				tt.mutate(&plan)

				return plan
			}())
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var got clusterModel
			require.False(t, resp.Plan.Get(context.Background(), &got).HasError())
			tt.check(t, got)
		})
	}
}

func TestClusterKeepStateModifiers(t *testing.T) {
	for _, tt := range []struct {
		name     string
		modify   planmodifier.String
		stateV   types.String
		configV  types.String
		plan     types.String
		expected types.String
	}{
		{name: "a case-only name change keeps the state name", modify: nameCase{}, stateV: types.StringValue("test-cluster"), configV: types.StringNull(), plan: types.StringValue("Test-Cluster"), expected: types.StringValue("test-cluster")},
		{name: "a case-only name change against another name is not rewritten", modify: nameCase{}, stateV: types.StringValue("test-cluster"), configV: types.StringNull(), plan: types.StringValue("other-cluster"), expected: types.StringValue("other-cluster")},
		{name: "a dash-only project change keeps the state project", modify: sameProjectID{}, stateV: types.StringValue("a07abc12310546f1b9291ab301a7d75b"), configV: types.StringValue("a07abc12-3105-46f1-b929-1ab301a7d75b"), plan: types.StringValue("a07abc12-3105-46f1-b929-1ab301a7d75b"), expected: types.StringValue("a07abc12310546f1b9291ab301a7d75b")},
		{name: "another project is not rewritten", modify: sameProjectID{}, stateV: types.StringValue("a07abc12310546f1b9291ab301a7d75b"), configV: types.StringValue("c07abc12310546f1b9291ab301a7d75b"), plan: types.StringValue("c07abc12310546f1b9291ab301a7d75b"), expected: types.StringValue("c07abc12310546f1b9291ab301a7d75b")},
		{name: "a null state project keeps the plan", modify: sameProjectID{}, stateV: types.StringNull(), configV: types.StringValue("x"), plan: types.StringValue("x"), expected: types.StringValue("x")},
		{name: "whitespace-only ca_certs keeps the state value", modify: trimmedStringKeep{}, stateV: types.StringValue("cert-1"), configV: types.StringValue("cert-1\n"), plan: types.StringValue("cert-1\n"), expected: types.StringValue("cert-1")},
		{name: "another ca_certs value is not rewritten", modify: trimmedStringKeep{}, stateV: types.StringValue("cert-1"), configV: types.StringValue("cert-2"), plan: types.StringValue("cert-2"), expected: types.StringValue("cert-2")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := planmodifier.StringResponse{PlanValue: tt.plan}
			tt.modify.PlanModifyString(context.Background(), planmodifier.StringRequest{
				StateValue: tt.stateV, ConfigValue: tt.configV, PlanValue: tt.plan,
			}, &resp)
			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}

func TestClusterAuditSecretModifierRules(t *testing.T) {
	r := &clusterResource{}
	reqOf := func(t *testing.T, planned, configAudit types.Object) planmodifier.StringRequest {
		t.Helper()
		plannedModel := clusterPlan()
		plannedModel.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"audit_logs": planned})
		configModel := plannedModel
		configModel.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"audit_logs": configAudit})
		st := clusterStateOf(t, r, plannedModel)

		return planmodifier.StringRequest{
			Config: tfsdk.Config(clusterStateOf(t, r, configModel)),
			Plan:   tfsdk.Plan(st),
			State:  st,
			Path:   path.Root("kubernetes_options").AtName("audit_logs").AtName("secret_name"),
		}
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
		{name: "a non-empty name with audit logs planned disabled fails the plan", planned: auditLogsObject(types.BoolValue(false), types.StringNull()), config: auditLogsObject(types.BoolValue(true), types.StringValue("siem")), configV: types.StringValue("siem"), wantErr: "Audit logs secret name set while audit logs are disabled"},
		{name: "planned disabled with a null config name plans empty", planned: auditLogsObject(types.BoolValue(false), types.StringNull()), config: auditLogsObject(types.BoolValue(false), types.StringNull()), configV: types.StringNull(), want: types.StringValue("")},
		{name: "a null name with audit logs planned enabled plans the API default", planned: auditLogsObject(types.BoolValue(true), types.StringNull()), config: auditLogsObject(types.BoolValue(true), types.StringNull()), configV: types.StringNull(), want: types.StringValue("mks-audit-logs")},
		{name: "planned unknown enabled with a null config name plans unknown", planned: auditLogsObject(types.BoolUnknown(), types.StringNull()), config: auditLogsObject(types.BoolNull(), types.StringNull()), configV: types.StringNull(), want: types.StringUnknown()},
		{name: "omitted enabled on an enabled state keeps the state name", planned: auditLogsObject(types.BoolValue(true), types.StringValue("siem")), config: auditLogsObject(types.BoolNull(), types.StringNull()), configV: types.StringNull(), stateV: types.StringValue("siem"), want: types.StringValue("siem")},
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
	oidcOf := func(enabled types.Bool, claim types.String) types.Object {
		return oidcObjectOf(map[string]attr.Value{"enabled": enabled, "username_claim": claim})
	}
	reqAt := func(t *testing.T, r *clusterResource, config, state clusterModel, field string) planmodifier.StringRequest {
		t.Helper()

		return planmodifier.StringRequest{
			Config:      tfsdk.Config(clusterStateOf(t, r, config)),
			State:       clusterStateOf(t, r, state),
			ConfigValue: types.StringNull(),
			StateValue:  types.StringValue("https://issuer.example.com"),
			Path:        path.Root("kubernetes_options").AtName("oidc").AtName(field),
		}
	}
	t.Run("an OIDC parameter set while OIDC is disabled fails the plan", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{
			"oidc": oidcObjectOf(map[string]attr.Value{"provider_name": types.StringValue("keycloak")}),
		})
		req := reqAt(t, r, config, clusterState(), "provider_name")
		req.ConfigValue = types.StringValue("keycloak")
		req.StateValue = types.StringNull()
		resp := planmodifier.StringResponse{PlanValue: req.ConfigValue}
		oidcPlanModifier{}.PlanModifyString(context.Background(), req, &resp)
		require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, "OIDC parameter set while OIDC is disabled", resp.Diagnostics.Errors()[0].Summary())
	})
	t.Run("an empty claim while enabled fails the plan", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidcOf(types.BoolValue(true), types.StringValue(""))})
		req := reqAt(t, r, config, clusterState(), "username_claim")
		req.ConfigValue = types.StringValue("")
		req.StateValue = types.StringNull()
		resp := planmodifier.StringResponse{PlanValue: types.StringValue("")}
		oidcPlanModifier{enabledDefault: "sub"}.PlanModifyString(context.Background(), req, &resp)
		require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, "Empty OIDC claim while OIDC is enabled", resp.Diagnostics.Errors()[0].Summary())
		assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "username_claim cannot be empty while OIDC is enabled: omit it to get")
	})
	t.Run("a null claim while enabled plans the OIDC default", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidcOf(types.BoolValue(true), types.StringNull())})
		state := clusterState()
		state.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": enabledOIDCObject()})
		req := reqAt(t, r, config, state, "username_claim")
		req.ConfigValue = types.StringNull()
		req.StateValue = types.StringNull()
		resp := planmodifier.StringResponse{PlanValue: types.StringUnknown()}
		oidcPlanModifier{enabledDefault: "sub"}.PlanModifyString(context.Background(), req, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, types.StringValue("sub"), resp.PlanValue)
	})
	t.Run("null enabled with an enabled state keeps the plan parameter", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidcOf(types.BoolNull(), types.StringValue("https://issuer.example.com"))})
		state := clusterState()
		state.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": enabledOIDCObject()})
		req := reqAt(t, r, config, state, "issuer_url")
		req.ConfigValue = types.StringValue("https://issuer.example.com")
		resp := planmodifier.StringResponse{PlanValue: req.ConfigValue}
		oidcPlanModifier{}.PlanModifyString(context.Background(), req, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, types.StringValue("https://issuer.example.com"), resp.PlanValue)
	})
	t.Run("null enabled on create plans an empty parameter", func(t *testing.T) {
		config := clusterPlan()
		config.KubernetesOptions = kubeOptionsObject(map[string]attr.Value{"oidc": oidcOf(types.BoolNull(), types.StringNull())})
		req := reqAt(t, r, config, clusterState(), "issuer_url")
		req.ConfigValue = types.StringNull()
		req.StateValue = types.StringNull()
		resp := planmodifier.StringResponse{PlanValue: types.StringUnknown()}
		oidcPlanModifier{}.PlanModifyString(context.Background(), req, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		assert.Equal(t, types.StringValue(""), resp.PlanValue)
	})
}
