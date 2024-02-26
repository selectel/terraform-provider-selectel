package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/mutexkv"
)

const (
	objectACL                     = "acl"
	objectFloatingIP              = "floating IP"
	objectKeypair                 = "keypair"
	objectLicense                 = "license"
	objectProject                 = "project"
	objectProjectQuotas           = "quotas for project"
	objectRole                    = "role"
	objectSubnet                  = "subnet"
	objectToken                   = "token"
	objectTopic                   = "topic"
	objectUser                    = "user"
	objectCluster                 = "cluster"
	objectKubeConfig              = "kubeconfig"
	objectKubeVersions            = "kube-versions"
	objectNodegroup               = "nodegroup"
	objectDomain                  = "domain"
	objectRecord                  = "record"
	objectZone                    = "zone"
	objectRRSet                   = "rrset"
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
	objectLogicalReplicationSlot  = "logical-replication-slot"
	objectRegistry                = "registry"
	objectRegistryToken           = "registry token"
)

// This is a global MutexKV for use within this plugin.
var selMutexKV = mutexkv.NewMutexKV()

// Provider returns the Selectel terraform provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
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
			"auth_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_AUTH_URL", nil),
				Description: "Base url to work with auth API (Keystone URL). https://api.selvpc.ru/identity/v3/ used by default",
			},
			"auth_region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_REGION_NAME", nil),
				Description: "Region for Keystone and Resell API URLs, 'ru-1' is used by default.",
			},
			"domain_name": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_DOMAIN_NAME", nil),
				Description: "Your domain name i.e. your account id",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USERNAME", nil),
				Description: "Service user username",
			},
			"user_domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USER_DOMAIN_NAME", nil),
				Description: "Used for service accounts in other domain. If you don't know exactly what this field means then don't use it",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_PASSWORD", nil),
				Description: "Service user password",
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
			"selectel_vpc_floatingip_v2":                            resourceVPCFloatingIPV2(),
			"selectel_vpc_keypair_v2":                               resourceVPCKeypairV2(),
			"selectel_vpc_license_v2":                               resourceVPCLicenseV2(),
			"selectel_vpc_project_v2":                               resourceVPCProjectV2(),
			"selectel_vpc_role_v2":                                  resourceVPCRoleV2(),
			"selectel_vpc_subnet_v2":                                resourceVPCSubnetV2(),
			"selectel_vpc_token_v2":                                 resourceVPCTokenV2(), // DEPRECATED
			"selectel_vpc_user_v2":                                  resourceVPCUserV2(),
			"selectel_vpc_vrrp_subnet_v2":                           resourceVPCVRRPSubnetV2(),        // DEPRECATED
			"selectel_vpc_crossregion_subnet_v2":                    resourceVPCCrossRegionSubnetV2(), // DEPRECATED
			"selectel_mks_cluster_v1":                               resourceMKSClusterV1(),
			"selectel_mks_nodegroup_v1":                             resourceMKSNodegroupV1(),
			"selectel_domains_domain_v1":                            resourceDomainsDomainV1(),
			"selectel_domains_record_v1":                            resourceDomainsRecordV1(),
			"selectel_domains_zone_v2":                              resourceDomainsZoneV2(),
			"selectel_domains_rrset_v2":                             resourceDomainsRRSetV2(),
			"selectel_dbaas_datastore_v1":                           resourceDBaaSDatastoreV1(), // DEPRECATED
			"selectel_dbaas_postgresql_datastore_v1":                resourceDBaaSPostgreSQLDatastoreV1(),
			"selectel_dbaas_mysql_datastore_v1":                     resourceDBaaSMySQLDatastoreV1(),
			"selectel_dbaas_redis_datastore_v1":                     resourceDBaaSRedisDatastoreV1(),
			"selectel_dbaas_user_v1":                                resourceDBaaSUserV1(),
			"selectel_dbaas_database_v1":                            resourceDBaaSDatabaseV1(), // DEPRECATED
			"selectel_dbaas_postgresql_database_v1":                 resourceDBaaSPostgreSQLDatabaseV1(),
			"selectel_dbaas_mysql_database_v1":                      resourceDBaaSMySQLDatabaseV1(),
			"selectel_dbaas_grant_v1":                               resourceDBaaSGrantV1(),
			"selectel_dbaas_extension_v1":                           resourceDBaaSExtensionV1(), // DEPRECATED
			"selectel_dbaas_postgresql_extension_v1":                resourceDBaaSPostgreSQLExtensionV1(),
			"selectel_dbaas_prometheus_metric_token_v1":             resourceDBaaSPrometheusMetricTokenV1(),
			"selectel_dbaas_postgresql_logical_replication_slot_v1": resourceDBaaSPostgreSQLLogicalReplicationSlotV1(),
			"selectel_dbaas_kafka_acl_v1":                           resourceDBaaSKafkaACLV1(),
			"selectel_dbaas_kafka_datastore_v1":                     resourceDBaaSKafkaDatastoreV1(),
			"selectel_dbaas_kafka_topic_v1":                         resourceDBaaSKafkaTopicV1(),
			"selectel_craas_registry_v1":                            resourceCRaaSRegistryV1(),
			"selectel_craas_token_v1":                               resourceCRaaSTokenV1(),
		},
		ConfigureContextFunc: configureProvider,
	}
}

func configureProvider(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config, diagError := getConfig(d)
	if diagError != nil {
		return nil, diagError
	}

	return config, nil
}
