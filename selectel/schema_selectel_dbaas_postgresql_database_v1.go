package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDBaaSPostgreSQLDatabaseV1Schema() map[string]*schema.Schema {
	databaseSchema := resourceDBaaSDatabaseV1BaseSchema()
	databaseSchema["owner_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}
	databaseSchema["lc_collate"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: true,
		Default:  "C",
	}
	databaseSchema["lc_ctype"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: true,
		Default:  "C",
	}

	return databaseSchema
}
