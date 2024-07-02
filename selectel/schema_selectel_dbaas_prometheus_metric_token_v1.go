package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceDBaaSPostgreSQLPrometheusMetricTokenV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"value": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}
