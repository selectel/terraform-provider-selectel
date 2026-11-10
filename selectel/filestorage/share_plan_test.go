package filestorage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShareAttributeValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*shareModel)
		valid  bool
	}{
		{"valid", func(*shareModel) {}, true},
		{"CIFS", func(m *shareModel) { m.ShareProto = types.StringValue(protoCIFS) }, true},
		{"unsupported protocol", func(m *shareModel) { m.ShareProto = types.StringValue("GLUSTERFS") }, false},
		{"lowercase protocol", func(m *shareModel) { m.ShareProto = types.StringValue("nfs") }, false},
		{"zero size", func(m *shareModel) { m.Size = types.Int64Value(0) }, false},
		{"ip omitted", func(m *shareModel) { m.IP = types.StringNull() }, true},
		{"ip is a hostname", func(m *shareModel) { m.IP = types.StringValue("storage.local") }, false},
		{"ip is IPv6", func(m *shareModel) { m.IP = types.StringValue("fd00::5") }, false},
		{"ip is a CIDR", func(m *shareModel) { m.IP = types.StringValue("10.0.0.5/24") }, false},
		{"ip out of range", func(m *shareModel) { m.IP = types.StringValue("10.0.0.256") }, false},
		{"network is not a UUID", func(m *shareModel) { m.NetworkID = types.StringValue("net") }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := testShareResource(nil)
			config := shareState()
			tc.change(&config)
			var s resource.SchemaResponse
			r.Schema(context.Background(), resource.SchemaRequest{}, &s)
			diags := attributeErrors(t, s.Schema.Attributes, tfsdk.Config(shareStateOf(t, r, config)))

			assert.Equal(t, !tc.valid, diags.HasError(), "%v", diags)
		})
	}
}

func TestShareConfigValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		change    func(*shareModel)
		errorText string
		detail    string
	}{
		{"valid", func(*shareModel) {}, "", ""},
		{"client metadata", func(m *shareModel) { m.Metadata = metadata(map[string]string{"owner": "team"}) }, "", ""},
		{"unknown type", func(m *shareModel) { m.ShareType = types.StringUnknown() }, "", ""},
		{"type UUID", func(m *shareModel) { m.ShareType = types.StringValue(networkID) }, "Share type UUID is not supported", "basic.ru-6a"},
		{"service key", func(m *shareModel) {
			m.Metadata = metadata(map[string]string{"delete_empty_share_network": "disabled"})
		}, "Reserved metadata key", "access_list_allow_all"},
		{"allow all key", func(m *shareModel) {
			m.Metadata = metadata(map[string]string{"access_list_allow_all": "true"})
		}, "Reserved metadata key", `access_to = "0.0.0.0/0"`},
		{"creation token", func(m *shareModel) { m.Metadata = metadata(map[string]string{createTokenKey: "x"}) }, "Reserved metadata key", ""},
		{"ip key", func(m *shareModel) { m.Metadata = metadata(map[string]string{"ip": "10.0.0.6"}) }, "Reserved metadata key", "use the ip argument"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := testShareResource(nil)
			config := shareState()
			tc.change(&config)
			response := resource.ValidateConfigResponse{}

			r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{
				Config: tfsdk.Config(shareStateOf(t, r, config)),
			}, &response)

			if tc.errorText == "" {
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
				return
			}
			require.Len(t, response.Diagnostics.Errors(), 1, "%v", response.Diagnostics)
			assert.Equal(t, tc.errorText, response.Diagnostics.Errors()[0].Summary())
			assert.Contains(t, response.Diagnostics.Errors()[0].Detail(), tc.detail)
		})
	}
}

func TestSharePlanSize(t *testing.T) {
	smaller := func(m *shareModel) { m.Size = types.Int64Value(40) }
	for name, tc := range map[string]struct {
		change          func(*shareModel)
		create, destroy bool
		err             string
	}{
		"unchanged":           {change: func(*shareModel) {}},
		"grow":                {change: func(m *shareModel) { m.Size = types.Int64Value(60) }},
		"shrink":              {change: smaller, err: "File storage shrinking is unsupported"},
		"unknown size":        {change: func(m *shareModel) { m.Size = types.Int64Unknown() }},
		"replace smaller":     {change: func(m *shareModel) { smaller(m); m.ShareType = types.StringValue("fast.ru-9a") }},
		"new address smaller": {change: func(m *shareModel) { smaller(m); m.IP = types.StringValue("10.0.0.6") }},
		"create":              {change: smaller, create: true},
		"destroy":             {change: smaller, destroy: true},
	} {
		t.Run(name, func(t *testing.T) {
			r := testShareResource(nil)
			prior := shareState()
			plan := prior
			tc.change(&plan)
			req := resource.ModifyPlanRequest{
				State: shareStateOf(t, r, prior),
				Plan:  tfsdk.Plan(shareStateOf(t, r, plan)),
			}
			if tc.create {
				req.State.Raw = tftypes.NewValue(req.State.Raw.Type(), nil)
			}
			if tc.destroy {
				req.Plan.Raw = tftypes.NewValue(req.Plan.Raw.Type(), nil)
			}
			resp := resource.ModifyPlanResponse{Plan: req.Plan}

			r.ModifyPlan(context.Background(), req, &resp)

			if tc.err == "" {
				assert.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				return
			}
			require.Len(t, resp.Diagnostics.Errors(), 1, "%v", resp.Diagnostics)
			assert.Equal(t, tc.err, resp.Diagnostics.Errors()[0].Summary())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "Current size is 50 GB")
		})
	}
}

// TestSharePlanReplacement plans through the framework server, which runs the
// attribute plan modifiers the way Terraform does: any change marks omitted
// computed values unknown before the modifiers see them.
func TestSharePlanReplacement(t *testing.T) {
	ctx := context.Background()
	server := providerserver.NewProtocol6(planTestProvider{})()
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	valueType := schemas.ResourceSchemas["selectel_filestorage_share_v2"].ValueType()
	r := testShareResource(nil)
	encode := func(m shareModel) *tfprotov6.DynamicValue {
		value, err := tfprotov6.NewDynamicValue(valueType, shareStateOf(t, r, m).Raw)
		require.NoError(t, err)

		return &value
	}
	replaces := func(name string) []*tftypes.AttributePath {
		return []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName(name)}
	}
	for _, tc := range []struct {
		name     string
		priorIP  string
		configIP types.String
		change   func(*shareModel)
		replace  []*tftypes.AttributePath
		ip       string
	}{
		{"rename", "10.0.0.5", types.StringNull(), func(m *shareModel) { m.Name = types.StringValue("renamed") }, nil, "10.0.0.5"},
		{"grow", "10.0.0.5", types.StringNull(), func(m *shareModel) { m.Size = types.Int64Value(60) }, nil, "10.0.0.5"},
		{"ip omitted", "10.0.0.5", types.StringNull(), func(*shareModel) {}, nil, "10.0.0.5"},
		{"ip equals the address", "10.0.0.5", types.StringValue("10.0.0.5"), func(*shareModel) {}, nil, "10.0.0.5"},
		{"ip differs from the address", "10.0.0.5", types.StringValue("10.0.0.6"), func(*shareModel) {}, replaces("ip"), "10.0.0.6"},
		{"protocol", "10.0.0.5", types.StringNull(), func(m *shareModel) { m.ShareProto = types.StringValue(protoCIFS) }, replaces("share_proto"), ""},
		{"type", "10.0.0.5", types.StringNull(), func(m *shareModel) { m.ShareType = types.StringValue("fast.ru-9a") }, replaces("share_type"), ""},
		{"network", "10.0.0.5", types.StringNull(), func(m *shareModel) { m.NetworkID = types.StringValue(shareID) }, replaces("network_id"), ""},
		{"subnet", "10.0.0.5", types.StringNull(), func(m *shareModel) { m.SubnetID = types.StringValue(shareID) }, replaces("subnet_id"), ""},
		{"project", "10.0.0.5", types.StringNull(), func(m *shareModel) { m.ProjectID = types.StringValue("other") }, replaces("project_id"), ""},
		{"region", "10.0.0.5", types.StringNull(), func(m *shareModel) { m.Region = types.StringValue("ru-1") }, replaces("region"), ""},
		{"ip set after an import without the address", "", types.StringValue("10.0.0.6"), func(*shareModel) {}, nil, "10.0.0.6"},
		{"ip omitted after an import without the address", "", types.StringNull(), func(*shareModel) {}, nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prior := shareState()
			prior.IP = types.StringValue(tc.priorIP)
			prior.Description = types.StringValue("set in the control panel")
			// Terraform proposes the configured values and keeps the prior ones
			// of computed attributes that the configuration omits.
			config := prior
			tc.change(&config)
			config.ID = types.StringNull()
			config.IP = tc.configIP
			config.AllMetadata = types.MapNull(types.StringType)
			// An omitted name and description, as after an import, keep the
			// values of the share.
			if config.Name.Equal(prior.Name) {
				config.Name = types.StringNull()
			}
			config.Description = types.StringNull()
			config.Status = types.StringNull()
			config.AvailabilityZone = types.StringNull()
			config.ExportLocations = types.ListNull(types.ObjectType{AttrTypes: exportLocationTypes})
			proposed := prior
			tc.change(&proposed)
			if !tc.configIP.IsNull() {
				proposed.IP = tc.configIP
			}

			resp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName:         "selectel_filestorage_share_v2",
				PriorState:       encode(prior),
				ProposedNewState: encode(proposed),
				Config:           encode(config),
			})
			require.NoError(t, err)
			require.Empty(t, resp.Diagnostics)
			assert.ElementsMatch(t, tc.replace, resp.RequiresReplace)
			if tc.replace == nil {
				planned, err := resp.PlannedState.Unmarshal(valueType)
				require.NoError(t, err)
				var attrs map[string]tftypes.Value
				require.NoError(t, planned.As(&attrs))
				var ip string
				require.NoError(t, attrs["ip"].As(&ip))
				assert.Equal(t, tc.ip, ip)
				var name, description string
				require.NoError(t, attrs["name"].As(&name))
				require.NoError(t, attrs["description"].As(&description))
				proposedName := proposed.Name.ValueString()
				assert.Equal(t, proposedName, name)
				assert.Equal(t, prior.Description.ValueString(), description)
			}
		})
	}
}

type planTestProvider struct{}

func (planTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "selectel"
}

func (planTestProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {}

func (planTestProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}

func (planTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewShareResource}
}

func (planTestProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }
