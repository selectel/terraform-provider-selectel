package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDBaaSDatastoreV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"subnet_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"type_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"flavor_id": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      false,
			ConflictsWith: []string{"flavor"},
		},
		"node_count": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"enabled": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"backup_retention_days": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Number of days to retain backups.",
		},
		"connections": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"flavor": {
			Type:          schema.TypeSet,
			Optional:      true,
			Computed:      true,
			ForceNew:      false,
			ConflictsWith: []string{"flavor_id"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"vcpus": {
						Type:     schema.TypeInt,
						Required: true,
					},
					"ram": {
						Type:     schema.TypeInt,
						Required: true,
					},
					"disk": {
						Type:     schema.TypeInt,
						Required: true,
					},
				},
			},
		},
		"pooler": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"mode": {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							"session",
							"transaction",
							"statement",
						}, false),
					},
					"size": {
						Type:     schema.TypeInt,
						Required: true,
					},
				},
			},
		},
		"firewall": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"ips": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"restore": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"datastore_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"target_time": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"config": {
			Type:     schema.TypeMap,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"redis_password": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
