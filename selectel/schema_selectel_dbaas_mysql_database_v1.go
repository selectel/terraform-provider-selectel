package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceDBaaSMySQLDatabaseV1Schema() map[string]*schema.Schema {
	return resourceDBaaSDatabaseV1BaseSchema()
}
