package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func vpcComputedFixedIPsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"subnet_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"ip_address": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func vpcComputedRoutesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"destination_cidr": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"next_hop": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}
