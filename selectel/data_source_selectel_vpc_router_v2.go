package selectel

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemaconfig"
)

func dataSourceVPCRouterV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCRouterV2Read,
		Description: "Provides a cloud router. " +
			"For more information about cloud routers, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/cloud-routers/).\n\n" +
			upgradeGuideNote,
		Schema: map[string]*schema.Schema{
			"project_id": projectIDDataSourceSchema(),
			"region":     regionDataSourceSchema("router"),
			"router_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Unique identifier of the router.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the router.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Description of the router.",
			},
			"admin_state_up": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Specifies if the router is administratively up.",
			},
			"ha": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the router runs in high availability mode.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the router.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags of the router.",
			},
			"external_network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Unique identifier of the external network the gateway is attached to.",
			},
			"enable_snat": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if source NAT is enabled on the gateway.",
			},
			"external_fixed_ip": vpcComputedFixedIPsSchema("Gateway addresses of the router."),
			"routes":            vpcComputedRoutesSchema("Static routes of the router."),
			"availability_zone_hints": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Pool segments requested for allocating the router.",
			},
			"availability_zones": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Pool segments where the router is actually allocated.",
			},
			"all_tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All tags of the router.",
			},
			"blocked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the router is administratively blocked.",
			},
		},
	}
}

func dataSourceVPCRouterV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	filters := make(map[string][]string)
	for schemaKey, queryKey := range map[string]string{
		"router_id":   "id",
		"name":        "name",
		"description": "description",
	} {
		if v, ok := d.GetOk(schemaKey); ok {
			filters[queryKey] = []string{v.(string)}
		}
	}
	if schemaconfig.IsSet(d, "admin_state_up") {
		filters["admin_state_up"] = []string{strconv.FormatBool(d.Get("admin_state_up").(bool))}
	}
	if v, ok := d.GetOk("tags"); ok {
		filters["tags"] = []string{strings.Join(convert.ToStringSlice(v.(*schema.Set).List()), ",")}
	}

	found, err := router.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(vpcRouterV2Object, err))
	}

	switch len(found) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			vpcRouterV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(found), vpcRouterV2Object))
	}

	result := found[0]

	d.SetId(result.ID)
	d.Set("router_id", result.ID)
	flattenVPCRouterV2(d, &result)

	if err := d.Set("routes", flattenVPCRouterV2Routes(result.Routes)); err != nil {
		log.Print(errSettingComplexAttr("routes", err))
	}

	log.Printf("[DEBUG] found VPC router %s", result.ID)

	return nil
}

func flattenVPCRouterV2Routes(routes []router.Route) []any {
	result := make([]any, 0, len(routes))
	for _, route := range routes {
		result = append(result, map[string]any{
			"destination_cidr": route.Destination,
			"next_hop":         route.NextHop,
		})
	}

	return result
}
