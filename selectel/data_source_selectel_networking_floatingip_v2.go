package selectel

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
)

func dataSourceNetworkingFloatingIPV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkingFloatingIPV2Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"floatingip_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			// pool is the name of the external network, resolved to its
			// identifier before the search.
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"port_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"floating_network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"router_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"qos_policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkingFloatingIPV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	filters := make(map[string][]string)
	for schemaKey, queryKey := range map[string]string{
		"floatingip_id": "id",
		"address":       "floating_ip_address",
		"description":   "description",
		"port_id":       "port_id",
		"fixed_ip":      "fixed_ip_address",
		"status":        "status",
	} {
		if v, ok := d.GetOk(schemaKey); ok {
			filters[queryKey] = []string{v.(string)}
		}
	}
	if v, ok := d.GetOk("tags"); ok {
		filters["tags"] = []string{strings.Join(expandToStringSlice(v.(*schema.Set).List()), ",")}
	}
	if v, ok := d.GetOk("pool"); ok {
		networkID, err := resolveNetworkingFloatingIPV2Pool(ctx, client, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		filters["floating_network_id"] = []string{networkID}
	}

	found, err := floatingip.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(networkingFloatingIPV2Object, err))
	}

	result, err := networkingFloatingIPV2DataSourceResult(found)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(result.ID)
	d.Set("floatingip_id", result.ID)
	flattenNetworkingFloatingIPV2(d, result)

	log.Printf("[DEBUG] found networking floating ip %s", result.ID)

	return nil
}

func networkingFloatingIPV2DataSourceResult(found []floatingip.FloatingIP) (*floatingip.FloatingIP, error) {
	switch len(found) {
	case 0:
		return nil, fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			networkingFloatingIPV2Object)
	case 1:
		return &found[0], nil
	default:
		return nil, fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(found), networkingFloatingIPV2Object)
	}
}
