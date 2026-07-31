package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
)

const networkingRouterInterfaceV2Object = "networking router interface"

func resourceNetworkingRouterInterfaceV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingRouterInterfaceV2Create,
		ReadContext:   resourceNetworkingRouterInterfaceV2Read,
		UpdateContext: resourceNetworkingRouterInterfaceV2Update,
		DeleteContext: resourceNetworkingRouterInterfaceV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingRouterInterfaceV2ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
			// Exactly one of the two addresses the interface. The one left out
			// of the configuration is computed from the attached port.
			"subnet_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"port_id"},
			},
			"port_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"subnet_id"},
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceNetworkingRouterInterfaceV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	subnetID := d.Get("subnet_id").(string)
	portID := d.Get("port_id").(string)

	// ConflictsWith rejects the pair at plan time; the empty case has no schema
	// equivalent, so both are checked here before the router is touched.
	switch {
	case subnetID == "" && portID == "":
		return diag.FromErr(errors.New(
			"exactly one of subnet_id or port_id must be set to attach the router to a network"))
	case subnetID != "" && portID != "":
		return diag.FromErr(errors.New(
			"subnet_id and port_id are mutually exclusive, please set exactly one of them"))
	}

	routerID := d.Get("router_id").(string)

	selector := router.BySubnet(subnetID)
	if portID != "" {
		// Neutron requires the port to carry exactly one fixed IP; checking it
		// here keeps a rejected attach from leaving a half-created resource.
		found, err := port.Get(ctx, client, portID)
		if err != nil {
			return diag.FromErr(errGettingObject(networkingPortV2Object, portID, err))
		}
		if len(found.FixedIPs) != 1 {
			return diag.FromErr(fmt.Errorf(
				"port %s has %d fixed ips, a router interface requires exactly one",
				portID, len(found.FixedIPs)))
		}

		selector = router.ByPort(portID)
	}

	log.Printf("[DEBUG] attaching networking router %s interface", routerID)
	attached, err := router.AddInterface(ctx, client, routerID, selector)
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingRouterInterfaceV2Object, err))
	}

	// The identity is the port the router is attached through, whether the
	// caller supplied it or the API created it for the subnet.
	d.SetId(attached.PortID)
	d.Set("port_id", attached.PortID)
	if attached.SubnetID != "" {
		d.Set("subnet_id", attached.SubnetID)
	}

	return resourceNetworkingRouterInterfaceV2Read(ctx, d, meta)
}

func resourceNetworkingRouterInterfaceV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking router interface port %s", d.Id())
	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		// A detached interface deletes its port, so a missing port is the
		// ordinary external-deletion case.
		return networkingCheckDeleted(d, err, networkingRouterInterfaceV2Object)
	}

	d.Set("port_id", found.ID)
	if found.DeviceID != "" {
		d.Set("router_id", found.DeviceID)
	}

	if subnetID := networkingRouterInterfaceV2SubnetFromPort(found.FixedIPs); subnetID != "" {
		d.Set("subnet_id", subnetID)
	}

	return nil
}

// resourceNetworkingRouterInterfaceV2Update exists only so force_destroy can
// change without replacing the interface: the attribute steers this resource's
// own Delete behaviour and has no remote counterpart to update.
func resourceNetworkingRouterInterfaceV2Update(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceNetworkingRouterInterfaceV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	subnetID := d.Get("subnet_id").(string)

	if d.Get("force_destroy").(bool) && subnetID != "" {
		// Clear only the routes that point into the subnet being detached: the
		// conflict they cause is the one this resource is allowed to resolve.
		if err := clearNetworkingRouterInterfaceV2Routes(ctx, client, routerID, subnetID); err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[DEBUG] detaching networking router %s interface %s", routerID, d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, detachErr := router.RemoveInterface(ctx, client, routerID, router.ByPort(d.Id()))
		if detachErr == nil {
			return nil
		}

		// The interface may still be in use by something this resource does not
		// own, so a conflict is retried until the timeout rather than forced.
		if networkingIsConflict(detachErr) {
			return resource.RetryableError(detachErr)
		}

		return resource.NonRetryableError(detachErr)
	})
	if err != nil {
		if networkingIsNotFound(err) {
			// Already detached or the port is gone: nothing left to do.
			return nil
		}

		return diag.FromErr(errDeletingObject(networkingRouterInterfaceV2Object, d.Id(), err))
	}

	return nil
}

func resourceNetworkingRouterInterfaceV2ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return nil, err
	}

	// router_id is Required, so import has to resolve it from the port before
	// the resource can be read.
	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		return nil, errGettingObject(networkingRouterInterfaceV2Object, d.Id(), err)
	}

	if found.DeviceID == "" {
		return nil, fmt.Errorf(
			"port %s is not attached to a router, so it is not a router interface", d.Id())
	}

	d.Set("router_id", found.DeviceID)
	d.Set("port_id", found.ID)
	if subnetID := networkingRouterInterfaceV2SubnetFromPort(found.FixedIPs); subnetID != "" {
		d.Set("subnet_id", subnetID)
	}

	return []*schema.ResourceData{d}, nil
}

// networkingRouterInterfaceV2SubnetFromPort resolves the interface subnet from the
// port's fixed IPs. It reports an empty string unless there is exactly one: a
// dual-stack port belongs to two subnets and neither one is the answer.
func networkingRouterInterfaceV2SubnetFromPort(fixedIPs []port.FixedIP) string {
	if len(fixedIPs) != 1 {
		return ""
	}

	return fixedIPs[0].SubnetID
}

// clearNetworkingRouterInterfaceV2Routes removes the router's static routes whose
// next hop lies inside the subnet being detached. Every other route is kept:
// force_destroy resolves the conflict caused by this interface, not any other.
func clearNetworkingRouterInterfaceV2Routes(
	ctx context.Context, client *vpc.Client, routerID, subnetID string,
) error {
	attachedSubnet, err := subnet.Get(ctx, client, subnetID)
	if err != nil {
		return errGettingObject(networkingSubnetV2Object, subnetID, err)
	}

	_, cidr, err := net.ParseCIDR(attachedSubnet.CIDR)
	if err != nil {
		return fmt.Errorf("could not parse cidr %q of subnet %s: %w", attachedSubnet.CIDR, subnetID, err)
	}

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		return errGettingObject(networkingRouterV2Object, routerID, err)
	}

	kept := make([]router.Route, 0, len(found.Routes))
	for _, route := range found.Routes {
		nextHop := net.ParseIP(route.NextHop)
		if nextHop != nil && cidr.Contains(nextHop) {
			log.Printf("[DEBUG] removing route %s via %s from router %s to detach subnet %s",
				route.Destination, route.NextHop, routerID, subnetID)

			continue
		}

		kept = append(kept, route)
	}

	if len(kept) == len(found.Routes) {
		return nil
	}

	if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{Routes: &kept}); err != nil {
		return errUpdatingObject(networkingRouterV2Object, routerID, err)
	}

	return nil
}
