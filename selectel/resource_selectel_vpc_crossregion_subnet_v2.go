package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceVPCCrossRegionSubnetV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCCrossRegionSubnetV2Create,
		Read:   resourceVPCCrossRegionSubnetV2Read,
		Delete: resourceVPCCrossRegionSubnetV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"regions": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Set:      hashRegions,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vlan_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"servers": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      hashServers,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"subnets": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      hashSubnets,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"network_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"project_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vlan_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vtep_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceVPCCrossRegionSubnetV2Create(d *schema.ResourceData, meta interface{}) error {
	return errResourceDeprecated("selectel_vpc_crossregion_subnet_v2")
}

func resourceVPCCrossRegionSubnetV2Read(d *schema.ResourceData, meta interface{}) error {
	return errResourceDeprecated("selectel_vpc_crossregion_subnet_v2")
}

func resourceVPCCrossRegionSubnetV2Delete(d *schema.ResourceData, meta interface{}) error {
	return errResourceDeprecated("selectel_vpc_crossregion_subnet_v2")
}
