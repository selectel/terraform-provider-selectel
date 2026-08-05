package selectel

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
)

const (
	networkingSubnetRouteV2Object = "networking subnet route"

	// networkingSubnetRouteV2IDSeparator separates the subnet identifier from
	// the route pair in the resource identifier.
	networkingSubnetRouteV2IDSeparator = "-route-"
)

func resourceVPCNetworkSubnetRouteV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCNetworkSubnetRouteV2Create,
		ReadContext:   resourceVPCNetworkSubnetRouteV2Read,
		DeleteContext: resourceVPCNetworkSubnetRouteV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCNetworkSubnetRouteV2ImportState,
		},
		// The resource owns one entry of the subnet's host_routes list and has
		// no update path of its own, so every attribute replaces it.
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
			"subnet_id": {
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

func resourceVPCNetworkSubnetRouteV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	subnetID := d.Get("subnet_id").(string)
	route := subnet.HostRoute{
		Destination: d.Get("destination_cidr").(string),
		NextHop:     d.Get("next_hop").(string),
	}

	// Adding one entry is a read-modify-write of the whole host_routes list, so
	// concurrent routes on the same subnet must not interleave.
	selMutexKV.Lock(subnetID)
	defer selMutexKV.Unlock(subnetID)

	found, err := subnet.Get(ctx, client, subnetID)
	if err != nil {
		if networkingIsNotFound(err) {
			// The parent subnet is already gone, so the route cannot exist
			// either; leaving the identifier unset keeps it out of the state.
			log.Printf("[DEBUG] parent subnet %s is absent, skipping route creation", subnetID)

			return nil
		}

		return diag.FromErr(errGettingObject(networkingSubnetV2Object, subnetID, err))
	}

	if networkingSubnetRouteV2Index(found.HostRoutes, route) >= 0 {
		return diag.FromErr(fmt.Errorf(
			"%s with destination_cidr '%s' and next_hop '%s' already exists on subnet '%s'",
			networkingSubnetRouteV2Object, route.Destination, route.NextHop, subnetID))
	}

	routes := append(append([]subnet.HostRoute{}, found.HostRoutes...), route)

	log.Printf("[DEBUG] adding route to networking subnet %s", subnetID)
	if _, err := subnet.Update(ctx, client, subnetID, subnet.UpdateRequest{
		HostRoutes: &routes,
	}); err != nil {
		return diag.FromErr(errCreatingObject(networkingSubnetRouteV2Object, err))
	}

	d.SetId(networkingSubnetRouteV2ID(subnetID, route))

	return resourceVPCNetworkSubnetRouteV2Read(ctx, d, meta)
}

func resourceVPCNetworkSubnetRouteV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	subnetID := d.Get("subnet_id").(string)
	route := subnet.HostRoute{
		Destination: d.Get("destination_cidr").(string),
		NextHop:     d.Get("next_hop").(string),
	}

	found, err := subnet.Get(ctx, client, subnetID)
	if err != nil {
		// The route lives inside the subnet, so a subnet that no longer exists
		// takes the route with it.
		return networkingCheckDeleted(d, err, networkingSubnetV2Object)
	}

	// While the subnet exists, a missing route is an error rather than a silent
	// state removal: something changed host_routes outside this resource.
	if networkingSubnetRouteV2Index(found.HostRoutes, route) < 0 {
		return diag.FromErr(fmt.Errorf(
			"%s with destination_cidr '%s' and next_hop '%s' is no longer present on subnet '%s'",
			networkingSubnetRouteV2Object, route.Destination, route.NextHop, subnetID))
	}

	d.Set("subnet_id", subnetID)
	d.Set("destination_cidr", route.Destination)
	d.Set("next_hop", route.NextHop)

	return nil
}

func resourceVPCNetworkSubnetRouteV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	subnetID := d.Get("subnet_id").(string)
	route := subnet.HostRoute{
		Destination: d.Get("destination_cidr").(string),
		NextHop:     d.Get("next_hop").(string),
	}

	selMutexKV.Lock(subnetID)
	defer selMutexKV.Unlock(subnetID)

	found, err := subnet.Get(ctx, client, subnetID)
	if err != nil {
		if networkingIsNotFound(err) {
			// Deleting the subnet has already removed the route.
			return nil
		}

		return diag.FromErr(errGettingObject(networkingSubnetV2Object, subnetID, err))
	}

	index := networkingSubnetRouteV2Index(found.HostRoutes, route)
	if index < 0 {
		return diag.FromErr(fmt.Errorf(
			"error deleting %s: destination_cidr '%s' with next_hop '%s' is already absent from subnet '%s'",
			networkingSubnetRouteV2Object, route.Destination, route.NextHop, subnetID))
	}

	routes := make([]subnet.HostRoute, 0, len(found.HostRoutes)-1)
	routes = append(routes, found.HostRoutes[:index]...)
	routes = append(routes, found.HostRoutes[index+1:]...)

	log.Printf("[DEBUG] removing route from networking subnet %s", subnetID)
	if _, err := subnet.Update(ctx, client, subnetID, subnet.UpdateRequest{
		HostRoutes: &routes,
	}); err != nil {
		return diag.FromErr(errDeletingObject(networkingSubnetRouteV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCNetworkSubnetRouteV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	subnetID, route, err := parseNetworkingSubnetRouteV2ID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("subnet_id", subnetID)
	d.Set("destination_cidr", route.Destination)
	d.Set("next_hop", route.NextHop)

	return []*schema.ResourceData{d}, nil
}

// networkingSubnetRouteV2ID renders the composite resource identifier.
func networkingSubnetRouteV2ID(subnetID string, route subnet.HostRoute) string {
	return fmt.Sprintf("%s%s%s-%s",
		subnetID, networkingSubnetRouteV2IDSeparator, route.Destination, route.NextHop)
}

// parseNetworkingSubnetRouteV2ID splits a composite identifier of the form
// "<subnet_id>-route-<destination_cidr>-<next_hop>". The subnet identifier is
// everything before the first separator, and the remainder must split into
// exactly two parts on the dash.
func parseNetworkingSubnetRouteV2ID(id string) (string, subnet.HostRoute, error) {
	subnetID, remainder, found := strings.Cut(id, networkingSubnetRouteV2IDSeparator)
	if !found || subnetID == "" {
		return "", subnet.HostRoute{}, errParseNetworkingSubnetRouteV2ID(id)
	}

	parts := strings.Split(remainder, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", subnet.HostRoute{}, errParseNetworkingSubnetRouteV2ID(id)
	}

	return subnetID, subnet.HostRoute{Destination: parts[0], NextHop: parts[1]}, nil
}

func errParseNetworkingSubnetRouteV2ID(id string) error {
	return fmt.Errorf(
		"unable to parse %s ID '%s': expected the format <subnet_id>%s<destination_cidr>-<next_hop>",
		networkingSubnetRouteV2Object, id, networkingSubnetRouteV2IDSeparator)
}

// networkingSubnetRouteV2Index reports the position of an exact
// destination/next-hop pair in the list, or -1 when it is absent. A matching
// destination with a different next hop is a different route.
func networkingSubnetRouteV2Index(routes []subnet.HostRoute, route subnet.HostRoute) int {
	for i, candidate := range routes {
		if candidate.Destination == route.Destination && candidate.NextHop == route.NextHop {
			return i
		}
	}

	return -1
}
