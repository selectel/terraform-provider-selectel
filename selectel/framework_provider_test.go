package selectel

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testFrameworkProvider struct {
	*frameworkProvider
}

func (p testFrameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return append(p.frameworkProvider.DataSources(ctx), func() datasource.DataSource {
		return &muxTestDataSource{}
	})
}

type muxTestDataSource struct {
	config *Config
}

func (d *muxTestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mux_test"
}

func (d *muxTestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Attributes: map[string]dsschema.Attribute{
			"username": dsschema.StringAttribute{Computed: true},
		},
	}
}

func (d *muxTestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	d.config, _ = req.ProviderData.(*Config)
}

func (d *muxTestDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.config == nil || d.config != cfgSingletone {
		resp.Diagnostics.AddError("Config not shared", "The data source did not get the Config the SDKv2 provider uses.")

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), d.config.Username)...)
}

var testProviderEnv = []string{"OS_AUTH_URL", "OS_REGION_NAME", "OS_DOMAIN_NAME", "OS_USERNAME", "OS_PASSWORD"}

func setTestProviderEnv(t *testing.T) {
	t.Helper()

	for _, env := range testProviderEnv {
		if os.Getenv(env) == "" {
			t.Setenv(env, "test")
		}
	}
}

func TestProviderServerSchema(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T)
	}{
		{
			name:  "credentials in env",
			setup: setTestProviderEnv,
		},
		{
			name: "credentials in provider block",
			setup: func(t *testing.T) {
				t.Helper()
				for _, env := range testProviderEnv {
					t.Setenv(env, "")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			server, err := testAccProtoV6ProviderFactories["selectel"]()
			require.NoError(t, err)

			resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
			require.NoError(t, err)
			assert.Empty(t, resp.Diagnostics)
			assert.Contains(t, resp.ResourceSchemas, "selectel_vpc_project_v2")
			assert.Contains(t, resp.DataSourceSchemas, "selectel_mux_test")
		})
	}
}

func TestFrameworkProviderSharesConfig(t *testing.T) {
	setTestProviderEnv(t)
	ctx := context.Background()

	server, err := testAccProtoV6ProviderFactories["selectel"]()
	require.NoError(t, err)

	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)

	configured, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{Config: nullConfig(t, schemas.Provider)})
	require.NoError(t, err)
	require.Empty(t, configured.Diagnostics)

	dataSource := schemas.DataSourceSchemas["selectel_mux_test"]
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "selectel_mux_test",
		Config:   nullConfig(t, dataSource),
	})
	require.NoError(t, err)
	require.Empty(t, read.Diagnostics)

	state, err := read.State.Unmarshal(dataSource.ValueType())
	require.NoError(t, err)
	var attrs map[string]tftypes.Value
	require.NoError(t, state.As(&attrs))
	var username string
	require.NoError(t, attrs["username"].As(&username))
	assert.Equal(t, os.Getenv("OS_USERNAME"), username)
}

func TestFrameworkProviderConfigureWithoutSDK(t *testing.T) {
	p := &frameworkProvider{sdk: Provider("test"), version: "test"}
	resp := &provider.ConfigureResponse{}

	p.Configure(context.Background(), provider.ConfigureRequest{}, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Nil(t, resp.ResourceData)
	assert.Nil(t, resp.DataSourceData)
}

// nullConfig is a configuration that sets none of the schema's attributes.
func nullConfig(t *testing.T, s *tfprotov6.Schema) *tfprotov6.DynamicValue {
	t.Helper()

	typ, ok := s.ValueType().(tftypes.Object)
	require.True(t, ok)
	attrs := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	value, err := tfprotov6.NewDynamicValue(typ, tftypes.NewValue(typ, attrs))
	require.NoError(t, err)

	return &value
}
