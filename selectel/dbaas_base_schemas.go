package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceDBaaSDatastoreV1BaseSchema() map[string]*schema.Schema {
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
			Deprecated: "firewall has been deprecated in favour of using `selectel_dbaas_firewall_v1` resource instead.",
		},
		"config": {
			Type:     schema.TypeMap,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"instances": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"role": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"floating_ip": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}
}

func resourceDBaaSDatabaseV1BaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
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
		"datastore_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}
