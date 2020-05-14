package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	objectFloatingIP    = "floating IP"
	objectKeypair       = "keypair"
	objectLicense       = "license"
	objectProject       = "project"
	objectProjectQuotas = "quotas for project"
	objectRole          = "role"
	objectSubnet        = "subnet"
	objectToken         = "token"
	objectUser          = "user"
	objectVRRPSubnet    = "VRRP subnet"
	objectCluster       = "cluster"
	objectNodegroup     = "nodegroup"
	objectDomain        = "domain"
	objectRecord        = "record"
)

// This is a global MutexKV for use within this plugin.
var selMutexKV = mutexkv.NewMutexKV()

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
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEL_PROJECT_ID", nil),
				Description: "VPC project ID to import resources that need the project scope auth token.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEL_REGION", nil),
				Description: "VPC region to import resources associated with the specific region.",
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
			"selectel_mks_nodegroup_v1":          resourceMKSNodegroupV1(),
			"selectel_domains_domain_v1":         resourceDomainsDomainV1(),
			"selectel_domains_record_v1":         resourceDomainsRecordV1(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Token:    d.Get("token").(string),
		Endpoint: d.Get("endpoint").(string),
	}
	if v, ok := d.GetOk("project_id"); ok {
		config.ProjectID = v.(string)
	}
	if v, ok := d.GetOk("region"); ok {
		config.Region = v.(string)
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}
