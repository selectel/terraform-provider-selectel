package selectel

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/router"
)

const (
	vpcRouterRouteV2Object      = "VPC router route"
	vpcRouterRouteV2IDSeparator = "-route-"
)

var routerRouteDocs = resourceDocs{Name: "route"}

func resourceVPCRouterRouteV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCRouterRouteV2Create,
		ReadContext:   resourceVPCRouterRouteV2Read,
		DeleteContext: resourceVPCRouterRouteV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCRouterRouteV2ImportState,
		},
		Description: "Creates and manages a single static route of a cloud router using public API v2. " +
			staticRoutesMoreInfo + "\n\n" +
			"Do not manage the same router with both resources: `selectel_vpc_router_route_v2` and `selectel_vpc_router_routes_v2`.\n\n" +
			upgradeGuideNote,
		Schema: routerRouteDocs.withDocsHints(map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Composite identifier of the route in the format `<router_id>-route-<destination_cidr>-<next_hop>`, " +
					"for example, `" + exampleResourceID + "-route-10.0.0.0/8-192.168.199.10`.",
			},
			"project_id": routerRouteDocs.projectIDResourceSchema(),
			"region":     routerRouteDocs.regionResourceSchema(),
			"router_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "Unique identifier of the router the route belongs to. " +
					routerIDFromResource,
			},
			"destination_cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
				Description:  "Destination CIDR of the route, for example, `10.0.0.0/8`.",
			},
			"next_hop": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPAddress,
				Description: "Next hop IP address for the route, for example, `192.168.199.10`. " +
					"Must be an address in a subnet connected to one of the router's internal interfaces. " +
					"Cannot be an address from the router's external gateway subnet.",
			},
		}),
	}
}

func resourceVPCRouterRouteV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	route := vpcRouterRouteV2FromResource(d)

	// Serialize read-modify-write updates per router.
	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		return diag.FromErr(errGettingObject(vpcRouterV2Object, routerID, err))
	}

	if slices.Contains(found.Routes, route) {
		return diag.FromErr(fmt.Errorf(
			"%s with destination_cidr '%s' and next_hop '%s' already exists on router '%s'",
			vpcRouterRouteV2Object, route.Destination, route.NextHop, routerID))
	}

	routes := append(append([]router.Route{}, found.Routes...), route)

	log.Printf("[DEBUG] adding route to VPC router %s", routerID)
	if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
		Routes: &routes,
	}); err != nil {
		return diag.FromErr(errCreatingObject(vpcRouterRouteV2Object, err))
	}

	d.SetId(vpcRouterRouteV2ID(routerID, route))

	return resourceVPCRouterRouteV2Read(ctx, d, meta)
}

func resourceVPCRouterRouteV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	route := vpcRouterRouteV2FromResource(d)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		return vpcCheckDeleted(d, err, vpcRouterV2Object)
	}

	if !slices.Contains(found.Routes, route) {
		log.Printf("[DEBUG] route %s via %s is gone from router %s, removing it from the state",
			route.Destination, route.NextHop, routerID)
		d.SetId("")

		return nil
	}

	d.Set("router_id", routerID)
	d.Set("destination_cidr", route.Destination)
	d.Set("next_hop", route.NextHop)

	return nil
}

func resourceVPCRouterRouteV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	route := vpcRouterRouteV2FromResource(d)

	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return nil
		}

		return diag.FromErr(errGettingObject(vpcRouterV2Object, routerID, err))
	}

	index := slices.Index(found.Routes, route)
	if index < 0 {
		return nil
	}

	routes := make([]router.Route, 0, len(found.Routes)-1)
	routes = append(routes, found.Routes[:index]...)
	routes = append(routes, found.Routes[index+1:]...)

	log.Printf("[DEBUG] removing route from VPC router %s", routerID)
	if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
		Routes: &routes,
	}); err != nil {
		return diag.FromErr(errDeletingObject(vpcRouterRouteV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCRouterRouteV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	routerID, route, err := parseVPCRouterRouteV2ID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("router_id", routerID)
	d.Set("destination_cidr", route.Destination)
	d.Set("next_hop", route.NextHop)

	return []*schema.ResourceData{d}, nil
}

func vpcRouterRouteV2FromResource(d *schema.ResourceData) router.Route {
	return router.Route{
		Destination: d.Get("destination_cidr").(string),
		NextHop:     d.Get("next_hop").(string),
	}
}

func vpcRouterRouteV2ID(routerID string, route router.Route) string {
	return fmt.Sprintf("%s%s%s-%s",
		routerID, vpcRouterRouteV2IDSeparator, route.Destination, route.NextHop)
}

func parseVPCRouterRouteV2ID(id string) (string, router.Route, error) {
	routerID, remainder, found := strings.Cut(id, vpcRouterRouteV2IDSeparator)
	if !found || routerID == "" {
		return "", router.Route{}, errParseVPCRouterRouteV2ID(id)
	}

	parts := strings.Split(remainder, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", router.Route{}, errParseVPCRouterRouteV2ID(id)
	}

	return routerID, router.Route{Destination: parts[0], NextHop: parts[1]}, nil
}

func errParseVPCRouterRouteV2ID(id string) error {
	return fmt.Errorf(
		"unable to parse %s ID '%s': expected the format <router_id>%s<destination_cidr>-<next_hop>",
		vpcRouterRouteV2Object, id, vpcRouterRouteV2IDSeparator)
}
