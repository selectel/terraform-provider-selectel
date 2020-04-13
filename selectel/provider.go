package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	objectFloatingIP        = "floating IP"
	objectKeypair           = "keypair"
	objectLicense           = "license"
	objectProject           = "project"
	objectProjectQuotas     = "quotas for project"
	objectRole              = "role"
	objectSubnet            = "subnet"
	objectToken             = "token"
	objectUser              = "user"
	objectVRRPSubnet        = "VRRP subnet"
	objectCluster           = "cluster"
	objectClusterNodegroups = "nodegroups of cluster"
)

// Provider returns the Selectel terraform provider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEL_TOKEN", nil),
				Description: "Token to authorize with the Selectel API.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEL_ENDPOINT", nil),
				Description: "Base endpoint to work with the Selectel API.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"selectel_vpc_floatingip_v2":         resourceVPCFloatingIPV2(),
			"selectel_vpc_keypair_v2":            resourceVPCKeypairV2(),
			"selectel_vpc_license_v2":            resourceVPCLicenseV2(),
			"selectel_vpc_project_v2":            resourceVPCProjectV2(),
			"selectel_vpc_role_v2":               resourceVPCRoleV2(),
			"selectel_vpc_subnet_v2":             resourceVPCSubnetV2(),
			"selectel_vpc_token_v2":              resourceVPCTokenV2(),
			"selectel_vpc_user_v2":               resourceVPCUserV2(),
			"selectel_vpc_vrrp_subnet_v2":        resourceVPCVRRPSubnetV2(),
			"selectel_vpc_crossregion_subnet_v2": resourceVPCCrossRegionSubnetV2(),
			"selectel_mks_cluster_v1":            resourceMKSClusterV1(),
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
