package selvpc

import "github.com/hashicorp/terraform/helper/schema"

func resourceResellFloatingIPV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"floating_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"fixed_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"servers": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}
