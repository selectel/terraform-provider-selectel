package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDBaaSV2ClickhouseShardGroupSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
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
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},

		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"shard_names": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			MinItems: 1,
			Required: true,
		},

		"description": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
