package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/mutexkv"
)

const (
	objectACL                       = "acl"
	objectFloatingIP                = "floating IP"
	objectKeypair                   = "keypair"
	objectLicense                   = "license"
	objectProject                   = "project"
	objectProjectQuotas             = "quotas for project"
	objectRole                      = "role"
	objectSubnet                    = "subnet"
	objectToken                     = "token"
	objectTopic                     = "topic"
	objectUser                      = "user"
	objectServiceUser               = "service user"
	objectS3Credentials             = "s3 credentials"
	objectSAMLFederation            = "saml federation"
	objectSAMLFederationCertificate = "saml federation certificate"
	objectGroup                     = "group"
	objectGroupMembership           = "group-membership"
	objectCluster                   = "cluster"
	objectKubeConfig                = "kubeconfig"
	objectKubeVersions              = "kube-versions"
	objectNodegroup                 = "nodegroup"
	objectDomain                    = "domain"
	objectRecord                    = "record"
	objectZone                      = "zone"
	objectRRSet                     = "rrset"
	objectDatastore                 = "datastore"
	objectDatabase                  = "database"
	objectGrant                     = "grant"
	objectExtension                 = "extension"
	objectDatastoreTypes            = "datastore-types"
	objectAvailableExtensions       = "available-extensions"
	objectFlavors                   = "flavors"
	objectConfigurationParameters   = "configuration-parameters"
	objectPrometheusMetricToken     = "prometheus-metric-token"
	objectFeatureGates              = "feature-gates"
	objectAdmissionControllers      = "admission-controllers"
	objectLogicalReplicationSlot    = "logical-replication-slot"
	objectRegistry                  = "registry"
	objectRegistryToken             = "registry token"
	objectSecret                    = "secret"
	objectCertificate               = "certificate"
	objectServer                    = "server"
	objectServerChip                = "server-chip"
	objectOS                        = "os"
	objectLocation                  = "location"
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
				DefaultFunc: schema.EnvDefaultFunc("INFRA_PROJECT_ID", nil),
				Description: "VPC project ID to import resources that need the project scope auth token.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFRA_REGION", nil),
				Description: "VPC region to import resources associated with the specific region.",
			},
			"auth_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_AUTH_URL", nil),
				Description: "Base url to work with auth API (Keystone URL).",
			},
			"auth_region": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_REGION_NAME", nil),
				Description: "Region for Keystone and Resell API URLs.",
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
			"selectel_domains_zone_v2":                  dataSourceDomainsZoneV2(),
			"selectel_domains_rrset_v2":                 dataSourceDomainsRRSetV2(),
			"selectel_dbaas_datastore_type_v1":          dataSourceDBaaSDatastoreTypeV1(),
			"selectel_dbaas_available_extension_v1":     dataSourceDBaaSAvailableExtensionV1(),
			"selectel_dbaas_flavor_v1":                  dataSourceDBaaSFlavorV1(),
			"selectel_dbaas_configuration_parameter_v1": dataSourceDBaaSConfigurationParameterV1(),
			"selectel_dbaas_prometheus_metric_token_v1": dataSourceDBaaSPrometheusMetricTokenV1(),
			"selectel_mks_kubeconfig_v1":                dataSourceMKSKubeconfigV1(),
			"selectel_mks_kube_versions_v1":             dataSourceMKSKubeVersionsV1(),
			"selectel_mks_feature_gates_v1":             dataSourceMKSFeatureGatesV1(),
			"selectel_mks_admission_controllers_v1":     dataSourceMKSAdmissionControllersV1(),
			"selectel_servers_configuration_v1":         dataSourceServersConfigurationV1(),
			"selectel_servers_os_v1":                    dataSourceServersOSV1(),
			"selectel_servers_location_v1":              dataSourceServersLocationV1(),
			"selectel_servers_public_subnet_v1":         dataSourceServersPublicSubnetV1(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"selectel_vpc_floatingip_v2":                            resourceVPCFloatingIPV2(),
			"selectel_vpc_keypair_v2":                               resourceVPCKeypairV2(),
			"selectel_vpc_license_v2":                               resourceVPCLicenseV2(),
			"selectel_vpc_project_v2":                               resourceVPCProjectV2(),
			"selectel_vpc_subnet_v2":                                resourceVPCSubnetV2(),
			"selectel_iam_serviceuser_v1":                           resourceIAMServiceUserV1(),
			"selectel_iam_user_v1":                                  resourceIAMUserV1(),
			"selectel_iam_s3_credentials_v1":                        resourceIAMS3CredentialsV1(),
			"selectel_iam_saml_federation_v1":                       resourceIAMSAMLFederationV1(),
			"selectel_iam_saml_federation_certificate_v1":           resourceIAMSAMLFederationCertificateV1(),
			"selectel_iam_group_v1":                                 resourceIAMGroupV1(),
			"selectel_iam_group_membership_v1":                      resourceIAMGroupMembershipV1(),
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
			"selectel_dbaas_firewall_v1":                            resourceDBaaSFirewallV1(),
			"selectel_craas_registry_v1":                            resourceCRaaSRegistryV1(),
			"selectel_craas_token_v1":                               resourceCRaaSTokenV1(),
			"selectel_secretsmanager_secret_v1":                     resourceSecretsManagerSecretV1(),
			"selectel_secretsmanager_certificate_v1":                resourceSecretsManagerCertificateV1(),
			"selectel_servers_server_v1":                            resourceServersServerV1(),
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
