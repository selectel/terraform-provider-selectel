package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceDBaaSRedisDatastoreV1Schema() map[string]*schema.Schema {
	datastoreSchema := resourceDBaaSDatastoreV1BaseSchema()
	datastoreSchema["backup_retention_days"] = &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "Number of days to retain backups.",
		Default:     7,
	}
	datastoreSchema["restore"] = &schema.Schema{
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
	}
	datastoreSchema["flavor"] = &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"vcpus": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"ram": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"disk": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			},
		},
	}
	datastoreSchema["flavor_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}
	datastoreSchema["redis_password"] = &schema.Schema{
		Type:      schema.TypeString,
		Optional:  true,
		Sensitive: true,
	}
	datastoreSchema["floating_ips"] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"master": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"replica": {
					Type:     schema.TypeInt,
					Required: true,
				},
			},
		},
	}

	return datastoreSchema
}
