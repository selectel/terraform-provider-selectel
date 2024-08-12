package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ResourceDBaaSFirewallV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"datastore_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"ips": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}
