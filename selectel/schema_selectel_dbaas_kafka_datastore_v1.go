package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var invalidKafkaFields = []string{
	"backup_retention_days",
	"pooler",
	"restore",
	"redis_password",
}

func resourceDBaaSKafkaDatastoreV1Schema() map[string]*schema.Schema {
	datastoreSchema := resourceDBaaSDatastoreV1Schema()
	for _, field := range invalidKafkaFields {
		delete(datastoreSchema, field)
	}
	return datastoreSchema
}
