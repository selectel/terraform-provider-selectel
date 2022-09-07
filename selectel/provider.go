package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/mutexkv"
)

const (
	objectFloatingIP              = "floating IP"
	objectKeypair                 = "keypair"
	objectLicense                 = "license"
	objectProject                 = "project"
	objectProjectQuotas           = "quotas for project"
	objectRole                    = "role"
	objectSubnet                  = "subnet"
	objectToken                   = "token"
	objectUser                    = "user"
	objectCluster                 = "cluster"
	objectKubeConfig              = "kubeconfig"
	objectKubeVersions            = "kube-versions"
	objectNodegroup               = "nodegroup"
	objectDomain                  = "domain"
	objectRecord                  = "record"
	objectDatastore               = "datastore"
	objectDatabase                = "database"
	objectGrant                   = "grant"
	objectExtension               = "extension"
	objectDatastoreTypes          = "datastore-types"
	objectAvailableExtensions     = "available-extensions"
	objectFlavors                 = "flavors"
	objectConfigurationParameters = "configuration-parameters"
	objectPrometheusMetricToken   = "prometheus-metric-token"
	objectFeatureGates            = "feature-gates"
	objectAdmissionControllers    = "admission-controllers"
)

// This is a global MutexKV for use within this plugin.
var selMutexKV = mutexkv.NewMutexKV()

// Provider returns the Selectel terraform provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEL_TOKEN", nil),
				Description: "Token to authorize with the Selectel API.",
			},
			"user": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"password", "domain_name"},
				DefaultFunc:  schema.EnvDefaultFunc("SEL_USER", nil),
				Description:  "Cloud user.",
			},
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				RequiredWith: []string{"user", "domain_name"},
				DefaultFunc:  schema.EnvDefaultFunc("SEL_PASSWORD", nil),
				Description:  "Cloud user password.",
			},
			"domain_name": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"user", "password"},
				DefaultFunc:  schema.EnvDefaultFunc("SEL_DOMAIN_NAME", nil),
				Description:  "Cloud domain ID to import resources that need the project scope auth token.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEL_ENDPOINT", nil),
				Description: "Base endpoint to work with the Selectel API.",
			},
			"auth_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_AUTH_URL", nil),
				Description: "Base endpoint to work with the Keystone API.",
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ru1Region,
					ru2Region,
					ru3Region,
					ru7Region,
					ru8Region,
					ru9Region,
				}, false),
				DefaultFunc: schema.EnvDefaultFunc("SEL_REGION", nil),
				Description: "Cloud region to import resources associated with the specific region.",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEL_PROJECT_ID", nil),
				Description: "Cloud project ID to import resources that need the project scope auth token.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"selectel_domains_domain_v1":                dataSourceDomainsDomainV1(),
			"selectel_dbaas_datastore_type_v1":          dataSourceDBaaSDatastoreTypeV1(),
			"selectel_dbaas_available_extension_v1":     dataSourceDBaaSAvailableExtensionV1(),
			"selectel_dbaas_flavor_v1":                  dataSourceDBaaSFlavorV1(),
			"selectel_dbaas_configuration_parameter_v1": dataSourceDBaaSConfigurationParameterV1(),
			"selectel_dbaas_prometheus_metric_token_v1": dataSourceDBaaSPrometheusMetricTokenV1(),
			"selectel_mks_kubeconfig_v1":                dataSourceMKSKubeconfigV1(),
			"selectel_mks_kube_versions_v1":             dataSourceMKSKubeVersionsV1(),
			"selectel_mks_feature_gates_v1":             dataSourceMKSFeatureGatesV1(),
			"selectel_mks_admission_controllers_v1":     dataSourceMKSAdmissionControllersV1(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"selectel_vpc_floatingip_v2":                resourceVPCFloatingIPV2(),
			"selectel_vpc_keypair_v2":                   resourceVPCKeypairV2(),
			"selectel_vpc_license_v2":                   resourceVPCLicenseV2(),
			"selectel_vpc_project_v2":                   resourceVPCProjectV2(),
			"selectel_vpc_role_v2":                      resourceVPCRoleV2(),
			"selectel_vpc_subnet_v2":                    resourceVPCSubnetV2(),
			"selectel_vpc_token_v2":                     resourceVPCTokenV2(),
			"selectel_vpc_user_v2":                      resourceVPCUserV2(),
			"selectel_vpc_vrrp_subnet_v2":               resourceVPCVRRPSubnetV2(),        // DEPRECATED
			"selectel_vpc_crossregion_subnet_v2":        resourceVPCCrossRegionSubnetV2(), // DEPRECATED
			"selectel_mks_cluster_v1":                   resourceMKSClusterV1(),
			"selectel_mks_nodegroup_v1":                 resourceMKSNodegroupV1(),
			"selectel_domains_domain_v1":                resourceDomainsDomainV1(),
			"selectel_domains_record_v1":                resourceDomainsRecordV1(),
			"selectel_dbaas_datastore_v1":               resourceDBaaSDatastoreV1(), // DEPRECATED
			"selectel_dbaas_postgresql_datastore_v1":    resourceDBaaSPostgreSQLDatastoreV1(),
			"selectel_dbaas_mysql_datastore_v1":         resourceDBaaSMySQLDatastoreV1(),
			"selectel_dbaas_redis_datastore_v1":         resourceDBaaSRedisDatastoreV1(),
			"selectel_dbaas_user_v1":                    resourceDBaaSUserV1(),
			"selectel_dbaas_database_v1":                resourceDBaaSDatabaseV1(), // DEPRECATED
			"selectel_dbaas_postgresql_database_v1":     resourceDBaaSPostgreSQLDatabaseV1(),
			"selectel_dbaas_mysql_database_v1":          resourceDBaaSMySQLDatabaseV1(),
			"selectel_dbaas_grant_v1":                   resourceDBaaSGrantV1(),
			"selectel_dbaas_extension_v1":               resourceDBaaSExtensionV1(), // DEPRECATED
			"selectel_dbaas_postgresql_extension_v1":    resourceDBaaSPostgreSQLExtensionV1(),
			"selectel_dbaas_prometheus_metric_token_v1": resourceDBaaSPrometheusMetricTokenV1(),
		},
		ConfigureContextFunc: configureProvider,
	}
}

func configureProvider(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var config Config
	if v, ok := d.GetOk("token"); ok {
		config.Token = v.(string)
	}
	if v, ok := d.GetOk("user"); ok {
		config.User = v.(string)
	}
	if v, ok := d.GetOk("password"); ok {
		config.Password = v.(string)
	}
	if v, ok := d.GetOk("endpoint"); ok {
		config.Endpoint = v.(string)
	}
	if v, ok := d.GetOk("auth_url"); ok {
		config.OSEndpoint = v.(string)
	}
	if v, ok := d.GetOk("domain_name"); ok {
		config.DomainName = v.(string)
	}
	if v, ok := d.GetOk("region"); ok {
		config.Region = v.(string)
	}
	if err := config.Validate(); err != nil {
		return nil, diag.FromErr(err)
	}

	return &config, nil
}
