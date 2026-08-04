package selectel

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
)

func dataSourceNetworkingSubnetV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkingSubnetV2Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cidr": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"gateway_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ip_version": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.IntInSlice([]int{
					networkingSubnetV2IPVersion4,
					networkingSubnetV2IPVersion6,
				}),
			},
			"dhcp_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ipv6_address_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(networkingSubnetV2IPv6Modes, false),
			},
			"ipv6_ra_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(networkingSubnetV2IPv6Modes, false),
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"enable_dhcp": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"dns_nameservers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"allocation_pools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			// host_routes reports every route of the subnet, including the ones
			// added by the separate subnet route resource.
			"host_routes": {
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
			},
			"service_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkingSubnetV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	filters := make(map[string][]string)
	for schemaKey, queryKey := range map[string]string{
		"subnet_id":         "id",
		"name":              "name",
		"description":       "description",
		"network_id":        "network_id",
		"cidr":              "cidr",
		"gateway_ip":        "gateway_ip",
		"ipv6_address_mode": "ipv6_address_mode",
		"ipv6_ra_mode":      "ipv6_ra_mode",
	} {
		if v, ok := d.GetOk(schemaKey); ok {
			filters[queryKey] = []string{v.(string)}
		}
	}
	if v, ok := d.GetOk("ip_version"); ok {
		filters["ip_version"] = []string{strconv.Itoa(v.(int))}
	}
	if v, ok := d.GetOkExists("dhcp_enabled"); ok {
		filters["enable_dhcp"] = []string{strconv.FormatBool(v.(bool))}
	}
	if v, ok := d.GetOk("tags"); ok {
		filters["tags"] = []string{strings.Join(expandToStringSlice(v.(*schema.Set).List()), ",")}
	}

	found, err := subnet.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(networkingSubnetV2Object, err))
	}

	switch len(found) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			networkingSubnetV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(found), networkingSubnetV2Object))
	}

	result := found[0]

	d.SetId(result.ID)
	d.Set("subnet_id", result.ID)
	d.Set("name", result.Name)
	d.Set("description", result.Description)
	d.Set("network_id", result.NetworkID)
	d.Set("cidr", result.CIDR)
	d.Set("gateway_ip", derefString(result.GatewayIP))
	d.Set("ip_version", result.IPVersion)
	d.Set("enable_dhcp", result.EnableDHCP)
	d.Set("ipv6_address_mode", derefString(result.IPv6AddressMode))
	d.Set("ipv6_ra_mode", derefString(result.IPv6RAMode))
	d.Set("blocked", result.Blocked)

	if err := d.Set("dns_nameservers", result.DNSNameservers); err != nil {
		log.Print(errSettingComplexAttr("dns_nameservers", err))
	}
	if err := d.Set("allocation_pools", flattenNetworkingSubnetV2AllocationPools(result.AllocationPools)); err != nil {
		log.Print(errSettingComplexAttr("allocation_pools", err))
	}
	if err := d.Set("host_routes", flattenNetworkingSubnetV2HostRoutes(result.HostRoutes)); err != nil {
		log.Print(errSettingComplexAttr("host_routes", err))
	}
	if err := d.Set("service_types", result.ServiceTypes); err != nil {
		log.Print(errSettingComplexAttr("service_types", err))
	}
	if err := d.Set("all_tags", result.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}

	return nil
}
