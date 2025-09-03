package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceDBaaSKafkaDatastoreV1Schema() map[string]*schema.Schema {
	datastoreSchema := resourceDBaaSDatastoreV1BaseSchema()
	datastoreSchema["log_platform"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Name of Log Platform group.",
	}

	return datastoreSchema
}
