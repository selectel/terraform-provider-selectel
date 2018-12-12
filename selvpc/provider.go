package selvpc

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns the selvpc terraform provider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEL_TOKEN", nil),
				Description: "Token to authorize with the Selectel VPC API.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEL_ENDPOINT", nil),
				Description: "Base endpoint to work with the Selectel VPC API.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"selvpc_resell_floatingip_v2": resourceResellFloatingIPV2(),
			"selvpc_resell_license_v2":    resourceResellLicenseV2(),
			"selvpc_resell_project_v2":    resourceResellProjectV2(),
			"selvpc_resell_role_v2":       resourceResellRoleV2(),
			"selvpc_resell_subnet_v2":     resourceResellSubnetV2(),
			"selvpc_resell_token_v2":      resourceResellTokenV2(),
			"selvpc_resell_user_v2":       resourceResellUserV2(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Token:    d.Get("token").(string),
		Endpoint: d.Get("endpoint").(string),
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &config, nil
}
