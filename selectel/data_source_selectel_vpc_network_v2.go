package selectel

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
)

func dataSourceVPCNetworkV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCNetworkV2Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"network_id": {
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
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"external": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// matching_subnet_cidr selects the network that owns a subnet with
			// this CIDR.
			"matching_subnet_cidr": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_state_up": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"dns_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_hints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"port_security_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"qos_policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_network_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_physical_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_segmentation_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_dns_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCNetworkV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
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
		filters["tags"] = []string{strings.Join(expandToStringSlice(v.(*schema.Set).List()), ",")}
	}

	found, err := network.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(networkingNetworkV2Object, err))
	}

	// The remaining selectors are not list filters, so they are applied to the
	// returned set.
	matched := make([]network.Network, 0, len(found))
	for _, candidate := range found {
		if v, ok := d.GetOkExists("external"); ok && candidate.RouterExternal != v.(bool) {
			continue
		}
		if v, ok := d.GetOk("mtu"); ok && candidate.MTU != v.(int) {
			continue
		}

		matched = append(matched, candidate)
	}

	if v, ok := d.GetOk("matching_subnet_cidr"); ok {
		matched, err = filterNetworkingNetworkV2BySubnetCIDR(ctx, client, matched, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	switch len(matched) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			networkingNetworkV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(matched), networkingNetworkV2Object))
	}

	result := matched[0]

	d.SetId(result.ID)
	d.Set("network_id", result.ID)
	flattenNetworkingNetworkV2(d, &result)

	if err := d.Set("subnets", result.Subnets); err != nil {
		log.Print(errSettingComplexAttr("subnets", err))
	}

	return nil
}

// filterNetworkingNetworkV2BySubnetCIDR keeps only the networks that have a
// subnet with the given CIDR attached.
func filterNetworkingNetworkV2BySubnetCIDR(
	ctx context.Context, client *vpc.Client, candidates []network.Network, cidr string,
) ([]network.Network, error) {
	subnets, err := subnet.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: map[string][]string{"cidr": {cidr}}},
	})
	if err != nil {
		return nil, errGettingObjects(networkingSubnetV2Object, err)
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
