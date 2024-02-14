package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var invalidMySQLFields = []string{
	"pooler",
	"redis_password",
}

func resourceDBaaSMySQLDatastoreV1Schema() map[string]*schema.Schema {
	datastoreSchema := resourceDBaaSDatastoreV1Schema()
	for _, field := range invalidMySQLFields {
		delete(datastoreSchema, field)
	}

	return datastoreSchema
}
