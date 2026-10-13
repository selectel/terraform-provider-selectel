package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/blockstorage"
)

// providerTypeName is the provider type name and the prefix of every
// framework resource and data source name.
const providerTypeName = "selectel"

// frameworkProvider serves the terraform-plugin-framework resources next to
// the SDKv2 provider.
type frameworkProvider struct {
	sdk     *sdkschema.Provider
	version string
}

var _ provider.Provider = &frameworkProvider{}

// ProviderServer returns the SDKv2 provider, upgraded to protocol 6, muxed
// with the framework provider.
func ProviderServer(ctx context.Context, providerVersion string) (func() tfprotov6.ProviderServer, error) {
	sdk := Provider(providerVersion)

	return muxProviderServer(ctx, sdk, &frameworkProvider{sdk: sdk, version: providerVersion})
}

func muxProviderServer(ctx context.Context, sdk *sdkschema.Provider, fw provider.Provider) (func() tfprotov6.ProviderServer, error) {
	upgradedSDK, err := tf5to6server.UpgradeServer(ctx, sdk.GRPCProvider)
	if err != nil {
		return nil, err
	}

	mux, err := tf6muxserver.NewMuxServer(ctx,
		func() tfprotov6.ProviderServer { return upgradedSDK },
		providerserver.NewProtocol6(fw),
	)
	if err != nil {
		return nil, err
	}

	return mux.ProviderServer, nil
}

func (p *frameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerTypeName
	resp.Version = p.version
}

// Schema copies the SDKv2 provider schema, because tf6muxserver requires the
// two to be identical. SDKv2 reports Required + EnvDefaultFunc as optional
// while the variable is set, so the copy does the same.
func (p *frameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	attrs := make(map[string]providerschema.Attribute, len(p.sdk.Schema))
	for name, s := range p.sdk.Schema {
		required := s.Required && sdkDefault(s) == ""
		attrs[name] = providerschema.StringAttribute{
			Required:    required,
			Optional:    !required,
			Sensitive:   s.Sensitive,
			Description: s.Description,
		}
	}

	resp.Schema = providerschema.Schema{Attributes: attrs}
}

func (p *frameworkProvider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// tf6muxserver configures servers in order and stops at the first error,
	// so the SDKv2 server listed first has already built the Config.
	config, ok := p.sdk.Meta().(*Config)
	if !ok {
		resp.Diagnostics.AddError("SDKv2 provider is not configured",
			"The framework provider uses the Config of the SDKv2 provider, which must be configured first.")

		return
	}

	resp.ResourceData = config
	resp.DataSourceData = config
}

func (p *frameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{blockstorage.NewVolumeResource}
}

func (p *frameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		blockstorage.NewVolumeDataSource,
		blockstorage.NewSnapshotDataSource,
		blockstorage.NewVolumeTypeDataSource,
	}
}

func sdkDefault(s *sdkschema.Schema) string {
	v, _ := s.DefaultValue()
	str, _ := v.(string)

	return str
}
