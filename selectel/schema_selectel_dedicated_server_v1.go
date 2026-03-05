package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
	dedicatedServerSchemaKeyDiskConfig               = "disk_config"
	dedicatedServerSchemaKeyName                     = "name"
	dedicatedServerSchemaKeyLevel                    = "level"
	dedicatedServerSchemaKeyDiskType                 = "disk_type"
	dedicatedServerSchemaKeyDiskCount                = "count"
	dedicatedServerSchemaKeyDiskName                 = "disk_name"
	dedicatedServerSchemaKeyMount                    = "mount"
	dedicatedServerSchemaKeySize                     = "size"
	dedicatedServerSchemaKeySizePercent              = "size_percent"
	dedicatedServerSchemaKeyRaid                     = "raid"
	dedicatedServerSchemaKeyFSType                   = "fs_type"
	dedicatedServerSchemaKeyOSPassword               = "os_password"
	dedicatedServerSchemaForceUpdateAdditionalParams = "force_update_additional_params"
	dedicatedServerSchemaKeyPowerState               = "power_state"

	dedicatedServerPowerStateOn      = "on"
	dedicatedServerPowerStateOff     = "off"
	dedicatedServerPowerActionReboot = "reboot"
)

func resourceDedicatedServerV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// required params
		dedicatedServerSchemaKeyProjectID: {
			Type:     schema.TypeString,
			Required: true,
		},
		dedicatedServerSchemaKeyConfigurationID: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
		},
		dedicatedServerSchemaKeyLocationID: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
		},
		dedicatedServerSchemaKeyOSID: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
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
									ValidateFunc: validation.StringInSlice([]string{
										string(dedicatedServerRaid0Level),
										string(dedicatedServerRaid1Level),
										string(dedicatedServerRaid10Level),
									}, false),
								},
								dedicatedServerSchemaKeyDiskType: {
									Type:     schema.TypeString,
									Required: true,
								},
								dedicatedServerSchemaKeyDiskCount: {
									Type:     schema.TypeInt,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
					dedicatedServerSchemaKeyDiskPartitions: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								dedicatedServerSchemaKeyDiskName: {
									Type:     schema.TypeString,
									Optional: true,
								},
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
									Computed: true,
								},
								dedicatedServerSchemaKeyRaid: {
									Type:     schema.TypeString,
									Optional: true,
								},
								dedicatedServerSchemaKeyFSType: {
									Type:     schema.TypeString,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
					dedicatedServerSchemaKeyDiskConfig: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								dedicatedServerSchemaKeyName: {
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
				},
			},
		},

		// optional network params
		dedicatedServerSchemaKeyPublicSubnetID: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.IsUUID,
		},
		dedicatedServerSchemaKeyPublicSubnetIP: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.IsIPAddress,
		},
		dedicatedServerSchemaKeyPrivateSubnet: {
			Type:     schema.TypeString,
			Optional: true,
		},

		// optional power params
		dedicatedServerSchemaKeyPowerState: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ValidateFunc: validation.StringInSlice([]string{
				dedicatedServerPowerStateOn,
				dedicatedServerPowerStateOff,
				dedicatedServerPowerActionReboot,
			}, false),
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
