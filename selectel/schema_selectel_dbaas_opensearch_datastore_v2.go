package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	dbaas_v2_common "github.com/selectel/dbaas-go/v2/common"
	dbaas_v2_os "github.com/selectel/dbaas-go/v2/opensearch"
)

func resourceDBaaSV2OpensearchDatastoreSchema() map[string]*schema.Schema {
	datastoreSchema := resourceDBaaSV2DatastoreBaseSchema()

	datastoreSchema["password"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "Datastore password",
	}

	datastoreSchema["node_group"] = &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,

		Elem: &schema.Resource{
			Schema: dbaasV2OpensearchNodeGroupSchema(),
		},
	}

	datastoreSchema["security_groups"] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true, // When creating a cluster without groups, the API itself sets the default group..
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.IsUUID,
		},
		MinItems: 1,
	}

	datastoreSchema["log_platform"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,

		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"log_group": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}

	return datastoreSchema
}

func dbaasV2OpensearchNodeGroupSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},

		"name": {
			Type:     schema.TypeString,
			Required: true,
		},

		"role": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				string(dbaas_v2_os.NodeGroupRoleData),
				string(dbaas_v2_os.NodeGroupRoleManager),
				string(dbaas_v2_os.NodeGroupRoleDashboard),
			},
				false,
			),
		},

		"node_count": {
			Type:     schema.TypeInt,
			Required: true,
		},

		"has_public_ips": {
			Type:     schema.TypeBool,
			Optional: true,
		},

		"flavor": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,

			Elem: &schema.Resource{
				Schema: dbaasV2OpensearchFlavorSchema(),
			},
		},

		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"instances": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"ip": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"floating_ip": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"availability_zone": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"hostname": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}
}

func dbaasV2OpensearchFlavorSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{

		"id": {
			Type:     schema.TypeString,
			Optional: true,
		},

		"type": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				string(dbaas_v2_common.FlavorTypeFIXED),
				string(dbaas_v2_common.FlavorTypeFlexible),
			},
				false,
			),
		},

		"vcpus": {
			Type:     schema.TypeInt,
			Optional: true,
		},

		"ram": {
			Type:     schema.TypeInt,
			Optional: true,
		},

		"disk": {
			Type:     schema.TypeInt,
			Optional: true,
		},

		"disk_type": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringInSlice([]string{
				string(dbaas_v2_common.FlavorDiskLocal),
				string(dbaas_v2_common.FlavorDiskNetworkUltra),
			},
				false,
			),
		},
	}
}
