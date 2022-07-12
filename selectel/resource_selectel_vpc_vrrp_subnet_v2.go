package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/go-selvpcclient/selvpcclient"
)

func resourceVPCVRRPSubnetV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCVRRPSubnetV2Create,
		ReadContext:   resourceVPCVRRPSubnetV2Read,
		DeleteContext: resourceVPCVRRPSubnetV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"master_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"slave_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"prefix_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      29,
				ValidateFunc: validation.IntBetween(24, 29),
			},
			"ip_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  selvpcclient.IPv4,
				ValidateFunc: validation.StringInSlice([]string{
					string(selvpcclient.IPv4),
					string(selvpcclient.IPv6),
				}, false),
			},
			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnets": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      hashSubnets,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vlan_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"project_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vtep_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
		},
	}
}

func resourceVPCVRRPSubnetV2Create(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_vrrp_subnet_v2"))
}

func resourceVPCVRRPSubnetV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_vrrp_subnet_v2"))
}

func resourceVPCVRRPSubnetV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_vrrp_subnet_v2"))
}
