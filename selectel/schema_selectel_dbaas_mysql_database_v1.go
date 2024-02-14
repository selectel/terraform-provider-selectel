package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var invalidMySQLDatabaseFields = []string{
	"owner_id",
	"lc_ctype",
	"lc_collate",
}

func resourceDBaaSMySQLDatabaseV1Schema() map[string]*schema.Schema {
	databaseSchema := resourceDBaaSDatabaseV1Schema()
	for _, field := range invalidMySQLDatabaseFields {
		delete(databaseSchema, field)
	}

	return databaseSchema
}
