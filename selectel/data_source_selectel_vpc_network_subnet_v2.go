package selectel

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemaconfig"
)

var vpcNetworkSubnetDataSourceDocs = dataSourceDocs{Name: "subnet"}

func dataSourceVPCNetworkSubnetV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCNetworkSubnetV2Read,
		Description: "Provides a subnet of a private network. " +
			"For more information about private network subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/).\n\n" +
			upgradeGuideNote,
		Schema: map[string]*schema.Schema{
			"id":         vpcNetworkSubnetDataSourceDocs.idSchema(),
			"project_id": vpcNetworkSubnetDataSourceDocs.projectIDSchema(),
			"region":     vpcNetworkSubnetDataSourceDocs.regionSchema(),
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Unique identifier of the subnet. " +
					networkSubnetIDFromControlPanel,
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the subnet.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Description of the subnet.",
			},
			"network_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Unique identifier of the parent network. " +
					networkIDFromControlPanel,
			},
			"cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsCIDR,
				Description:  "CIDR of the subnet, for example `192.168.1.0/24`.",
			},
			"gateway_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "Gateway of the subnet, for example `192.168.1.1`.",
			},
			"ip_version": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.IntInSlice([]int{
					vpcNetworkSubnetV2IPVersion4,
					vpcNetworkSubnetV2IPVersion6,
				}),
				Description: "Internet protocol version. Available versions are `4` or `6`.",
			},
			"ipv6_address_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(vpcNetworkSubnetV2IPv6Modes, false),
				Description: "IPv6 address assignment mode. " +
					"Available modes are `slaac`, `dhcpv6-stateful`, or `dhcpv6-stateless`.",
			},
			"ipv6_ra_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(vpcNetworkSubnetV2IPv6Modes, false),
				Description: "IPv6 router advertisement mode. " +
					"Available modes are `slaac`, `dhcpv6-stateful`, or `dhcpv6-stateless`.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags of the subnet.",
			},
			"enable_dhcp": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Specifies if DHCP is enabled for the subnet.",
			},
			"dns_nameservers": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "DNS nameservers of the subnet.",
			},
			"allocation_pool": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "First address of the pool.",
						},
						"end": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last address of the pool.",
						},
					},
				},
				Description: "Allocation pools of the subnet.",
			},
			"host_routes": vpcComputedRoutesSchema(
				"Host routes of the subnet.",
			),
			"service_types": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Service types of the subnet.",
			},
			"all_tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All tags of the subnet.",
			},
			"blocked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the subnet is administratively blocked.",
			},
		},
	}
}

func dataSourceVPCNetworkSubnetV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
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
	if schemaconfig.IsSet(d, "enable_dhcp") {
		filters["enable_dhcp"] = []string{strconv.FormatBool(d.Get("enable_dhcp").(bool))}
	}
	if v, ok := d.GetOk("tags"); ok {
		filters["tags"] = []string{strings.Join(convert.ToStringSlice(v.(*schema.Set).List()), ",")}
	}

	found, err := subnet.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(vpcNetworkSubnetV2Object, err))
	}

	switch len(found) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			vpcNetworkSubnetV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(found), vpcNetworkSubnetV2Object))
	}

	result := found[0]

	d.SetId(result.ID)
	d.Set("subnet_id", result.ID)
	flattenVPCNetworkSubnetV2Common(d, &result)

	return nil
}
