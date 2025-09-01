package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	serversServerSchemaKeyProjectID                = "project_id"
	serversServerSchemaKeyConfigurationID          = "configuration_id"
	serversServerSchemaKeyLocationID               = "location_id"
	serversServerSchemaKeyOSID                     = "os_id"
	serversServerSchemaKeyPricePlanName            = "price_plan_name"
	serversServerSchemaKeyIsServerChip             = "is_server_chip"
	serversServerSchemaKeyPublicSubnetID           = "public_subnet_id"
	serversServerSchemaKeyPrivateSubnet            = "private_subnet"
	serversServerSchemaKeyOSUserData               = "user_data"
	serversServerSchemaKeyOSHostName               = "os_host_name"
	serversServerSchemaKeyOSSSHKey                 = "ssh_key"
	serversServerSchemaKeyOSSSHKeyName             = "ssh_key_name"
	serversServerSchemaKeyOSPartitionsConfig       = "partitions_config"
	serversServerSchemaKeySoftRaidConfig           = "soft_raid_config"
	serversServerSchemaKeyDiskPartitions           = "disk_partitions"
	serversServerSchemaKeyName                     = "name"
	serversServerSchemaKeyLevel                    = "level"
	serversServerSchemaKeyDiskType                 = "disk_type"
	serversServerSchemaKeyMount                    = "mount"
	serversServerSchemaKeySize                     = "size"
	serversServerSchemaKeySizePercent              = "size_percent"
	serversServerSchemaKeyRaid                     = "raid"
	serversServerSchemaKeyFSType                   = "fs_type"
	serversServerSchemaKeyOSPassword               = "os_password"
	serversServerSchemaForceUpdateAdditionalParams = "force_update_additional_params"
)

func resourceServersServerV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// required params
		serversServerSchemaKeyProjectID: {
			Type:     schema.TypeString,
			Required: true,
		},
		serversServerSchemaKeyConfigurationID: {
			Type:     schema.TypeString,
			Required: true,
		},
		serversServerSchemaKeyLocationID: {
			Type:     schema.TypeString,
			Required: true,
		},
		serversServerSchemaKeyOSID: {
			Type:     schema.TypeString,
			Required: true,
		},
		serversServerSchemaKeyPricePlanName: {
			Type:     schema.TypeString,
			Required: true,
		},

		// optional os params
		serversServerSchemaKeyOSPassword: {
			Type:     schema.TypeString,
			Optional: true,
		},
		serversServerSchemaKeyOSUserData: {
			Type:     schema.TypeString,
			Optional: true,
		},
		serversServerSchemaKeyOSSSHKey: {
			Type:     schema.TypeString,
			Optional: true,
		},
		serversServerSchemaKeyOSSSHKeyName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		serversServerSchemaKeyOSPartitionsConfig: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					serversServerSchemaKeySoftRaidConfig: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								serversServerSchemaKeyName: {
									Type:     schema.TypeString,
									Required: true,
								},
								serversServerSchemaKeyLevel: {
									Type:     schema.TypeString,
									Required: true,
								},
								serversServerSchemaKeyDiskType: {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					serversServerSchemaKeyDiskPartitions: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								serversServerSchemaKeyMount: {
									Type:     schema.TypeString,
									Required: true,
								},
								serversServerSchemaKeySize: {
									Type:     schema.TypeFloat,
									Optional: true,
								},
								serversServerSchemaKeySizePercent: {
									Type:     schema.TypeFloat,
									Optional: true,
								},
								serversServerSchemaKeyRaid: {
									Type:     schema.TypeString,
									Required: true,
								},
								serversServerSchemaKeyFSType: {
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
		serversServerSchemaKeyPublicSubnetID: {
			Type:     schema.TypeString,
			Optional: true,
		},
		serversServerSchemaKeyPrivateSubnet: {
			Type:     schema.TypeString,
			Optional: true,
		},

		// optional misc
		serversServerSchemaKeyIsServerChip: {
			Type:     schema.TypeBool,
			Optional: true,
		},
		serversServerSchemaKeyOSHostName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		serversServerSchemaForceUpdateAdditionalParams: {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
}
