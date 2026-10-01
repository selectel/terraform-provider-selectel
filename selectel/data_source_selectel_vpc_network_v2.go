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
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemaconfig"
)

func dataSourceVPCNetworkV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCNetworkV2Read,
		Description: "Provides a private network. " +
			"For more information about private networks, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/).\n\n" +
			upgradeGuideNote,
		Schema: map[string]*schema.Schema{
			"project_id": projectIDDataSourceSchema(),
			"region":     regionDataSourceSchema("network"),
			"network_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Unique identifier of the network. " +
					"To get the network ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ copy the ID of the required network. " +
					"As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack network list` and copy the `ID` field.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the network.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Description of the network.",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Status of the network. Available statuses are `ACTIVE`, `DOWN`, `BUILD`, or `ERROR`.",
			},
			"external": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Specifies if the network has an external routing facility.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum transmission unit of the network, for example `1500`.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags of the network.",
			},
			"matching_subnet_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsCIDR,
				Description:  "CIDR of a subnet within the network, for example `192.168.1.0/24`.",
			},
			"admin_state_up": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the network is administratively up.",
			},
			"dns_domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS domain of the network.",
			},
			"availability_zone_hints": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Pool segments requested for allocating the network.",
			},
			"availability_zones": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Pool segments where the network is actually allocated.",
			},
			"subnets": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Unique identifiers of the subnets within the network.",
			},
			"all_tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All tags of the network.",
			},
			"shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the network can be accessed by any project.",
			},
			"port_security_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if port security is enabled by default for the ports of the network.",
			},
			"blocked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the network is administratively blocked.",
			},
			"is_public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the network is public.",
			},
			"is_dns_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the Private DNS service is enabled for the network.",
			},
		},
	}
}

func dataSourceVPCNetworkV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	filters := make(map[string][]string)
	for schemaKey, queryKey := range map[string]string{
		"network_id":  "id",
		"name":        "name",
		"description": "description",
		"status":      "status",
	} {
		if v, ok := d.GetOk(schemaKey); ok {
			filters[queryKey] = []string{v.(string)}
		}
	}
	if v, ok := d.GetOk("tags"); ok {
		filters["tags"] = []string{strings.Join(convert.ToStringSlice(v.(*schema.Set).List()), ",")}
	}
	if schemaconfig.IsSet(d, "external") {
		filters["router:external"] = []string{strconv.FormatBool(d.Get("external").(bool))}
	}
	if v, ok := d.GetOk("mtu"); ok {
		filters["mtu"] = []string{strconv.Itoa(v.(int))}
	}

	matched, err := network.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(vpcNetworkV2Object, err))
	}

	if v, ok := d.GetOk("matching_subnet_cidr"); ok {
		matched, err = filterVPCNetworkV2BySubnetCIDR(ctx, client, matched, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	switch len(matched) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			vpcNetworkV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(matched), vpcNetworkV2Object))
	}

	result := matched[0]

	d.SetId(result.ID)
	d.Set("network_id", result.ID)
	flattenVPCNetworkV2(d, &result)

	return nil
}

func filterVPCNetworkV2BySubnetCIDR(
	ctx context.Context, client *vpc.Client, candidates []network.Network, cidr string,
) ([]network.Network, error) {
	subnets, err := subnet.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: map[string][]string{"cidr": {cidr}}},
	})
	if err != nil {
		return nil, errGettingObjects(vpcNetworkSubnetV2Object, err)
	}

	networkIDs := make(map[string]struct{}, len(subnets))
	for _, item := range subnets {
		networkIDs[item.NetworkID] = struct{}{}
	}

	filtered := make([]network.Network, 0, len(candidates))
	for _, candidate := range candidates {
		if _, ok := networkIDs[candidate.ID]; ok {
			filtered = append(filtered, candidate)
		}
	}

	return filtered, nil
}
