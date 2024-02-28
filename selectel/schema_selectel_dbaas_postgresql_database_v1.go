package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDBaaSPostgreSQLDatabaseV1Schema() map[string]*schema.Schema {
	databaseSchema := resourceDBaaSDatabaseV1BaseSchema()
	databaseSchema["owner"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}
	databaseSchema["lc_collate"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		ForceNew:         true,
		DiffSuppressFunc: dbaasDatabaseV1LocaleDiffSuppressFunc,
	}
	databaseSchema["lc_ctype"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		ForceNew:         true,
		DiffSuppressFunc: dbaasDatabaseV1LocaleDiffSuppressFunc,
	}

	return databaseSchema
}

func dbaasDatabaseV1LocaleDiffSuppressFunc(_, old, new string, _ *schema.ResourceData) bool {
	// The default locale value - C is the same as null value, so we need to suppress
	if old == "C" && new == "" {
		return true
	}

	return false
}
