package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceDBaaSKafkaDatastoreV1Schema() map[string]*schema.Schema {
	return resourceDBaaSDatastoreV1BaseSchema()
}
