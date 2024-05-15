package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDBaaSKafkaACKV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"datastore_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"pattern": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"pattern_type": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			ValidateFunc: validation.StringInSlice([]string{
				"literal",
				"prefixed",
				"all",
			}, false),
		},
		"user_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"allow_read": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"allow_write": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
	}
}
