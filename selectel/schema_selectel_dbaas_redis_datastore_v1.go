package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var invalidRedisFields = []string{
	"flavor",
	"flavor_id",
	"pooler",
}

func resourceDBaaSRedisDatastoreV1Schema() map[string]*schema.Schema {
	datastoreSchema := resourceDBaaSDatastoreV1Schema()
	for _, field := range invalidRedisFields {
		delete(datastoreSchema, field)
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

	return datastoreSchema
}
