package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	dedicatedServerSchemaKeyProjectID                = "project_id"
	dedicatedServerSchemaKeyConfigurationID          = "configuration_id"
	dedicatedServerSchemaKeyLocationID               = "location_id"
	dedicatedServerSchemaKeyOSID                     = "os_id"
	dedicatedServerSchemaKeyPricePlanName            = "price_plan_name"
	dedicatedServerSchemaKeyPublicSubnetID           = "public_subnet_id"
	dedicatedServerSchemaKeyPublicSubnetIP           = "public_subnet_ip"
	dedicatedServerSchemaKeyPrivateSubnet            = "private_subnet"
	dedicatedServerSchemaKeyOSUserData               = "user_data"
	dedicatedServerSchemaKeyOSHostName               = "os_host_name"
	dedicatedServerSchemaKeyOSSSHKey                 = "ssh_key"
	dedicatedServerSchemaKeyOSSSHKeyName             = "ssh_key_name"
	dedicatedServerSchemaKeyOSPartitionsConfig       = "partitions_config"
	dedicatedServerSchemaKeySoftRaidConfig           = "soft_raid_config"
	dedicatedServerSchemaKeyDiskPartitions           = "disk_partitions"
	dedicatedServerSchemaKeyName                     = "name"
	dedicatedServerSchemaKeyLevel                    = "level"
	dedicatedServerSchemaKeyDiskType                 = "disk_type"
	dedicatedServerSchemaKeyMount                    = "mount"
	dedicatedServerSchemaKeySize                     = "size"
	dedicatedServerSchemaKeySizePercent              = "size_percent"
	dedicatedServerSchemaKeyRaid                     = "raid"
	dedicatedServerSchemaKeyFSType                   = "fs_type"
	dedicatedServerSchemaKeyOSPassword               = "os_password"
	dedicatedServerSchemaForceUpdateAdditionalParams = "force_update_additional_params"
)

func resourceDedicatedServerV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// required params
		dedicatedServerSchemaKeyProjectID: {
			Type:     schema.TypeString,
			Required: true,
		},
		dedicatedServerSchemaKeyConfigurationID: {
			Type:     schema.TypeString,
			Required: true,
		},
		dedicatedServerSchemaKeyLocationID: {
			Type:     schema.TypeString,
			Required: true,
		},
		dedicatedServerSchemaKeyOSID: {
			Type:     schema.TypeString,
			Required: true,
		},
		dedicatedServerSchemaKeyPricePlanName: {
			Type:     schema.TypeString,
			Required: true,
		},

		// optional os params
		dedicatedServerSchemaKeyOSPassword: {
			Type:      schema.TypeString,
			Sensitive: true,
			Optional:  true,
		},
		dedicatedServerSchemaKeyOSUserData: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServerSchemaKeyOSSSHKey: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServerSchemaKeyOSSSHKeyName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServerSchemaKeyOSPartitionsConfig: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					dedicatedServerSchemaKeySoftRaidConfig: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								dedicatedServerSchemaKeyName: {
									Type:     schema.TypeString,
									Required: true,
								},
								dedicatedServerSchemaKeyLevel: {
									Type:     schema.TypeString,
									Required: true,
								},
								dedicatedServerSchemaKeyDiskType: {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					dedicatedServerSchemaKeyDiskPartitions: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								dedicatedServerSchemaKeyMount: {
									Type:     schema.TypeString,
									Required: true,
								},
								dedicatedServerSchemaKeySize: {
									Type:     schema.TypeFloat,
									Optional: true,
								},
								dedicatedServerSchemaKeySizePercent: {
									Type:     schema.TypeFloat,
									Optional: true,
								},
								dedicatedServerSchemaKeyRaid: {
									Type:     schema.TypeString,
									Required: true,
								},
								dedicatedServerSchemaKeyFSType: {
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
		dedicatedServerSchemaKeyPublicSubnetID: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServerSchemaKeyPublicSubnetIP: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServerSchemaKeyPrivateSubnet: {
			Type:     schema.TypeString,
			Optional: true,
		},

		// optional misc
		dedicatedServerSchemaKeyOSHostName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		dedicatedServerSchemaForceUpdateAdditionalParams: {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
}
