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
			"matching_subnet_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsCIDR,
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
			"availability_zones": {
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
