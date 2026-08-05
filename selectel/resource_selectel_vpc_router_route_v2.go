package selectel

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/vpc-go/pkg/v2/router"
)

const (
	networkingRouterRouteV2Object = "networking router route"

	// networkingRouterRouteV2IDSeparator separates the router identifier from the
	// route pair in the resource identifier.
	networkingRouterRouteV2IDSeparator = "-route-"
)

func resourceVPCRouterRouteV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCRouterRouteV2Create,
		ReadContext:   resourceVPCRouterRouteV2Read,
		DeleteContext: resourceVPCRouterRouteV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCRouterRouteV2ImportState,
		},
		// The resource owns one entry of the router's routes list and has no
		// update path of its own, so every attribute replaces it.
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"router_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"next_hop": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCRouterRouteV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	route := networkingRouterRouteV2FromResource(d)

	// Adding one entry is a read-modify-write of the whole routes list, so
	// concurrent routes on the same router must not interleave.
	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		if networkingIsNotFound(err) {
			// The parent router is already gone, so the route cannot exist
			// either; leaving the identifier unset keeps it out of the state.
			log.Printf("[DEBUG] parent router %s is absent, skipping route creation", routerID)

			return nil
		}

		return diag.FromErr(errGettingObject(networkingRouterV2Object, routerID, err))
	}

	if networkingRouterRouteV2Index(found.Routes, route) >= 0 {
		// An identical entry added outside this resource is an idempotent
		// success: the resource simply adopts it, and routes stays untouched.
		log.Printf("[DEBUG] route %s via %s already present on router %s, adopting it",
			route.Destination, route.NextHop, routerID)

		d.SetId(networkingRouterRouteV2ID(routerID, route))

		return resourceVPCRouterRouteV2Read(ctx, d, meta)
	}

	routes := append(append([]router.Route{}, found.Routes...), route)

	log.Printf("[DEBUG] adding route to networking router %s", routerID)
	if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
		Routes: &routes,
	}); err != nil {
		return diag.FromErr(errCreatingObject(networkingRouterRouteV2Object, err))
	}

	d.SetId(networkingRouterRouteV2ID(routerID, route))

	return resourceVPCRouterRouteV2Read(ctx, d, meta)
}

func resourceVPCRouterRouteV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	route := networkingRouterRouteV2FromResource(d)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		// The route lives inside the router, so a router that no longer exists
		// takes the route with it. Any other error keeps the resource in state.
		return networkingCheckDeleted(d, err, networkingRouterV2Object)
	}

	// The router is reachable, so a missing entry means the route itself was
	// removed outside this resource — the same observable outcome as the router
	// being deleted, and handled the same way.
	if networkingRouterRouteV2Index(found.Routes, route) < 0 {
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

func resourceVPCRouterRouteV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	route := networkingRouterRouteV2FromResource(d)

	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		if networkingIsNotFound(err) {
			// Deleting the router has already removed the route.
			return nil
		}

		return diag.FromErr(errGettingObject(networkingRouterV2Object, routerID, err))
	}

	index := networkingRouterRouteV2Index(found.Routes, route)
	if index < 0 {
		// Already absent: the outcome Delete wants is the outcome that holds, so
		// routes is not rewritten.
		return nil
	}

	routes := make([]router.Route, 0, len(found.Routes)-1)
	routes = append(routes, found.Routes[:index]...)
	routes = append(routes, found.Routes[index+1:]...)

	log.Printf("[DEBUG] removing route from networking router %s", routerID)
	if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
		Routes: &routes,
	}); err != nil {
		return diag.FromErr(errDeletingObject(networkingRouterRouteV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCRouterRouteV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	routerID, route, err := parseNetworkingRouterRouteV2ID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("router_id", routerID)
	d.Set("destination_cidr", route.Destination)
	d.Set("next_hop", route.NextHop)

	return []*schema.ResourceData{d}, nil
}

// networkingRouterRouteV2FromResource reads the route pair this resource owns.
func networkingRouterRouteV2FromResource(d *schema.ResourceData) router.Route {
	return router.Route{
		Destination: d.Get("destination_cidr").(string),
		NextHop:     d.Get("next_hop").(string),
	}
}

// networkingRouterRouteV2ID renders the composite resource identifier.
func networkingRouterRouteV2ID(routerID string, route router.Route) string {
	return fmt.Sprintf("%s%s%s-%s",
		routerID, networkingRouterRouteV2IDSeparator, route.Destination, route.NextHop)
}

// parseNetworkingRouterRouteV2ID splits a composite identifier of the form
// "<router_id>-route-<destination_cidr>-<next_hop>". The router identifier is
// everything before the first separator, and the remainder must split into
// exactly two parts on the dash — one destination can be served by several
// entries with different next hops, so only the pair identifies one route.
func parseNetworkingRouterRouteV2ID(id string) (string, router.Route, error) {
	routerID, remainder, found := strings.Cut(id, networkingRouterRouteV2IDSeparator)
	if !found || routerID == "" {
		return "", router.Route{}, errParseNetworkingRouterRouteV2ID(id)
	}

	parts := strings.Split(remainder, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", router.Route{}, errParseNetworkingRouterRouteV2ID(id)
	}

	return routerID, router.Route{Destination: parts[0], NextHop: parts[1]}, nil
}

func errParseNetworkingRouterRouteV2ID(id string) error {
	return fmt.Errorf(
		"unable to parse %s ID '%s': expected the format <router_id>%s<destination_cidr>-<next_hop>",
		networkingRouterRouteV2Object, id, networkingRouterRouteV2IDSeparator)
}

// networkingRouterRouteV2Index reports the position of an exact
// destination/next-hop pair in the list, or -1 when it is absent. A matching
// destination with a different next hop is a different route.
func networkingRouterRouteV2Index(routes []router.Route, route router.Route) int {
	for i, candidate := range routes {
		if candidate.Destination == route.Destination && candidate.NextHop == route.NextHop {
			return i
		}
	}

	return -1
}
