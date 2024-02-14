package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var invalidPostgreSQLFields = []string{
	"redis_password",
}

func resourceDBaaSPostgreSQLDatastoreV1Schema() map[string]*schema.Schema {
	datastoreSchema := resourceDBaaSDatastoreV1Schema()
	for _, field := range invalidPostgreSQLFields {
		delete(datastoreSchema, field)
	}
	return datastoreSchema
}
