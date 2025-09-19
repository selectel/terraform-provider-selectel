package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	dedicatedServersServerSchemaKeyProjectID                = "project_id"
	dedicatedServersServerSchemaKeyConfigurationID          = "configuration_id"
	dedicatedServersServerSchemaKeyLocationID               = "location_id"
	dedicatedServersServerSchemaKeyOSID                     = "os_id"
	dedicatedServersServerSchemaKeyPricePlanName            = "price_plan_name"
	dedicatedServersServerSchemaKeyPublicSubnetID           = "public_subnet_id"
	dedicatedServersServerSchemaKeyPrivateSubnet            = "private_subnet"
	dedicatedServersServerSchemaKeyOSUserData               = "user_data"
	dedicatedServersServerSchemaKeyOSHostName               = "os_host_name"
	dedicatedServersServerSchemaKeyOSSSHKey                 = "ssh_key"
	dedicatedServersServerSchemaKeyOSSSHKeyName             = "ssh_key_name"
	dedicatedServersServerSchemaKeyOSPartitionsConfig       = "partitions_config"
	dedicatedServersServerSchemaKeySoftRaidConfig           = "soft_raid_config"
	dedicatedServersServerSchemaKeyDiskPartitions           = "disk_partitions"
	dedicatedServersServerSchemaKeyName                     = "name"
	dedicatedServersServerSchemaKeyLevel                    = "level"
	dedicatedServersServerSchemaKeyDiskType                 = "disk_type"
	dedicatedServersServerSchemaKeyMount                    = "mount"
	dedicatedServersServerSchemaKeySize                     = "size"
	dedicatedServersServerSchemaKeySizePercent              = "size_percent"
	dedicatedServersServerSchemaKeyRaid                     = "raid"
	dedicatedServersServerSchemaKeyFSType                   = "fs_type"
	dedicatedServersServerSchemaKeyOSPassword               = "os_password"
	dedicatedServersServerSchemaForceUpdateAdditionalParams = "force_update_additional_params"
)

func resourceDedicatedServersServerV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// required params
		dedicatedServersServerSchemaKeyProjectID: {
			Type:     schema.TypeString,
			Required: true,
		},
		dedicatedServersServerSchemaKeyConfigurationID: {
			Type:     schema.TypeString,
			Required: true,
		},
		dedicatedServersServerSchemaKeyLocationID: {
			Type:     schema.TypeString,
			Required: true,
		},
		dedicatedServersServerSchemaKeyOSID: {
			Type:     schema.TypeString,
			Required: true,
		},
		dedicatedServersServerSchemaKeyPricePlanName: {
			Type:     schema.TypeString,
			Required: true,
		},

		// optional os params
		dedicatedServersServerSchemaKeyOSPassword: {
			Type:      schema.TypeString,
			Sensitive: true,
			Optional:  true,
		},
		dedicatedServersServerSchemaKeyOSUserData: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServersServerSchemaKeyOSSSHKey: {
			Type:      schema.TypeString,
			Sensitive: true,
			Optional:  true,
		},
		dedicatedServersServerSchemaKeyOSSSHKeyName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServersServerSchemaKeyOSPartitionsConfig: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					dedicatedServersServerSchemaKeySoftRaidConfig: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								dedicatedServersServerSchemaKeyName: {
									Type:     schema.TypeString,
									Required: true,
								},
								dedicatedServersServerSchemaKeyLevel: {
									Type:     schema.TypeString,
									Required: true,
								},
								dedicatedServersServerSchemaKeyDiskType: {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					dedicatedServersServerSchemaKeyDiskPartitions: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								dedicatedServersServerSchemaKeyMount: {
									Type:     schema.TypeString,
									Required: true,
								},
								dedicatedServersServerSchemaKeySize: {
									Type:     schema.TypeFloat,
									Optional: true,
								},
								dedicatedServersServerSchemaKeySizePercent: {
									Type:     schema.TypeFloat,
									Optional: true,
								},
								dedicatedServersServerSchemaKeyRaid: {
									Type:     schema.TypeString,
									Required: true,
								},
								dedicatedServersServerSchemaKeyFSType: {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
				},
			},
		},

		// optional network params
		dedicatedServersServerSchemaKeyPublicSubnetID: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServersServerSchemaKeyPrivateSubnet: {
			Type:     schema.TypeString,
			Optional: true,
		},

		// optional misc
		dedicatedServersServerSchemaKeyOSHostName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServersServerSchemaForceUpdateAdditionalParams: {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
}
