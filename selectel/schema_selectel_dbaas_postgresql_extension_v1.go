package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceDBaaSPostgreSQLExtensionV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		"available_extension_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"datastore_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"database_id": {
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
