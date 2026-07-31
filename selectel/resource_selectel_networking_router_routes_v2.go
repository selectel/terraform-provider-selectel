package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/vpc-go/pkg/v2/router"
)

const networkingRouterRoutesV2Object = "networking router routes"

func resourceNetworkingRouterRoutesV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingRouterRoutesV2Create,
		ReadContext:   resourceNetworkingRouterRoutesV2Read,
		UpdateContext: resourceNetworkingRouterRoutesV2Update,
		DeleteContext: resourceNetworkingRouterRoutesV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingRouterRoutesV2ImportState,
		},
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
			// routes is the authoritative declaration of the router's static
			// routes: Read fills it with the router's actual list, so anything
			// added by another mechanism shows up as drift and is planned away.
			"routes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_cidr": {
							Type:     schema.TypeString,
							Required: true,
						},
						"next_hop": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceNetworkingRouterRoutesV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	desired := expandNetworkingRouterRoutesV2(d.Get("routes"))

	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		return diag.FromErr(errGettingObject(networkingRouterV2Object, routerID, err))
	}

	// Create only ensures the declared routes are present. It does not remove
	// anything: the authoritative removal happens from the next Update on,
	// against a state that records what this resource observed.
	routes := append([]router.Route{}, found.Routes...)
	for _, route := range desired {
		if networkingRouterRouteV2Index(routes, route) < 0 {
			routes = append(routes, route)
		}
	}

	if len(routes) != len(found.Routes) {
		log.Printf("[DEBUG] adding %d route(s) to networking router %s", len(routes)-len(found.Routes), routerID)
		if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
			Routes: &routes,
		}); err != nil {
			return diag.FromErr(errCreatingObject(networkingRouterRoutesV2Object, err))
		}
	}

	d.SetId(routerID)

	return resourceNetworkingRouterRoutesV2Read(ctx, d, meta)
}

func resourceNetworkingRouterRoutesV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting routes of networking router %s", d.Id())
	found, err := router.Get(ctx, client, d.Id())
	if err != nil {
		return networkingCheckDeleted(d, err, networkingRouterV2Object)
	}

	d.Set("router_id", found.ID)
	if err := d.Set("routes", flattenNetworkingRouterRoutesV2(found.Routes)); err != nil {
		log.Print(errSettingComplexAttr("routes", err))
	}

	return nil
}

func resourceNetworkingRouterRoutesV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if !d.HasChange("routes") {
		return resourceNetworkingRouterRoutesV2Read(ctx, d, meta)
	}

	routerID := d.Get("router_id").(string)
	desired := expandNetworkingRouterRoutesV2(d.Get("routes"))

	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	// The configured list is authoritative, so it is sent as the router's whole
	// route list: anything not declared is removed, whoever added it.
	log.Printf("[DEBUG] replacing routes of networking router %s with %d entrie(s)", routerID, len(desired))
	if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
		Routes: &desired,
	}); err != nil {
		return diag.FromErr(errUpdatingObject(networkingRouterRoutesV2Object, routerID, err))
	}

	return resourceNetworkingRouterRoutesV2Read(ctx, d, meta)
}

func resourceNetworkingRouterRoutesV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	managed := expandNetworkingRouterRoutesV2(d.Get("routes"))
	if len(managed) == 0 {
		// Nothing was managed, so there is nothing to clear.
		return nil
	}

	routerID := d.Get("router_id").(string)

	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		if networkingIsNotFound(err) {
			// Deleting the router has already removed its routes.
			return nil
		}

		return diag.FromErr(errGettingObject(networkingRouterV2Object, routerID, err))
	}

	// Only the managed entries are removed; routes this resource never observed
	// stay on the router.
	kept := make([]router.Route, 0, len(found.Routes))
	for _, route := range found.Routes {
		if networkingRouterRouteV2Index(managed, route) < 0 {
			kept = append(kept, route)
		}
	}

	if len(kept) == len(found.Routes) {
		// Every managed entry is already absent.
		return nil
	}

	log.Printf("[DEBUG] removing %d managed route(s) from networking router %s", len(found.Routes)-len(kept), routerID)
	if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
		Routes: &kept,
	}); err != nil {
		return diag.FromErr(errDeletingObject(networkingRouterRoutesV2Object, routerID, err))
	}

	return nil
}

func resourceNetworkingRouterRoutesV2ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return nil, err
	}

	// A router_id that matches nothing must fail the import rather than produce
	// an empty resource in the state.
	found, err := router.Get(ctx, client, d.Id())
	if err != nil {
		return nil, errGettingObject(networkingRouterV2Object, d.Id(), err)
	}

	d.Set("router_id", found.ID)
	if err := d.Set("routes", flattenNetworkingRouterRoutesV2(found.Routes)); err != nil {
		log.Print(errSettingComplexAttr("routes", err))
	}

	return []*schema.ResourceData{d}, nil
}

// expandNetworkingRouterRoutesV2 reads the declared route set into SDK routes.
func expandNetworkingRouterRoutesV2(value interface{}) []router.Route {
	set, ok := value.(*schema.Set)
	if !ok {
		return []router.Route{}
	}

	routes := make([]router.Route, 0, set.Len())
	for _, raw := range set.List() {
		entry := raw.(map[string]interface{})
		routes = append(routes, router.Route{
			Destination: entry["destination_cidr"].(string),
			NextHop:     entry["next_hop"].(string),
		})
	}

	return routes
}

// flattenNetworkingRouterRoutesV2 writes SDK routes into the schema shape.
func flattenNetworkingRouterRoutesV2(routes []router.Route) []interface{} {
	result := make([]interface{}, 0, len(routes))
	for _, route := range routes {
		result = append(result, map[string]interface{}{
			"destination_cidr": route.Destination,
			"next_hop":         route.NextHop,
		})
	}

	return result
}
