package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
)

const (
	vpcRouterInterfaceV2Object = "VPC router interface"

	vpcRouterInterfaceV2InUseByFloatingIP = "RouterInterfaceInUseByFloatingIP"
)

var routerInterfaceDocs = resourceDocs{Name: "router interface"}

func resourceVPCRouterInterfaceV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCRouterInterfaceV2Create,
		ReadContext:   resourceVPCRouterInterfaceV2Read,
		UpdateContext: schema.NoopContext,
		DeleteContext: resourceVPCRouterInterfaceV2Delete,
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(vpcV2DefaultDeleteTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCRouterInterfaceV2ImportState,
		},
		Description: "Creates and manages an interface that attaches a cloud router to a network using public API v2. " +
			"For more information about cloud routers, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/cloud-routers/).\n\n" +
			upgradeGuideNote,
		Schema: routerInterfaceDocs.withDocsHints(map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Unique identifier of the port the router is attached through, for example, `" + exampleResourceID + "`.",
			},
			"project_id": projectIDResourceSchema(),
			"region":     routerInterfaceDocs.regionResourceSchema(),
			"router_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "Unique identifier of the router the network is attached to. " +
					"Retrieved from the [selectel_vpc_router_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2) resource.",
			},
			"network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Unique identifier of the network containing the interface port.",
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"subnet_id", "port_id"},
				Description: "Unique identifier of the subnet to attach. " +
					"Conflicts with `port_id`. " +
					"When attach through `subnet_id` creates the port for the interface. " +
					"Retrieved from the [selectel_vpc_network_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) resource.",
			},
			"port_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"subnet_id", "port_id"},
				Description: "Unique identifier of the port to attach. " +
					"Conflicts with `subnet_id`. " +
					"Port passed as `port_id` must have exactly one fixed IP. " +
					"Retrieved from the [selectel_vpc_port_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2) resource.",
			},
			"force_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies if static routes of the router that point into the attached subnet are deleted before the interface is detached.",
			},
		}),
	}
}

func resourceVPCRouterInterfaceV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	subnetID := d.Get("subnet_id").(string)
	portID := d.Get("port_id").(string)

	routerID := d.Get("router_id").(string)

	// Attaching mutates the router, so serialize updates per router.
	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	request := router.InterfaceRequest{SubnetID: subnetID}
	if portID != "" {
		found, err := port.Get(ctx, client, portID)
		if err != nil {
			return diag.FromErr(errGettingObject(vpcPortV2Object, portID, err))
		}
		if len(found.FixedIPs) != 1 {
			return diag.FromErr(fmt.Errorf(
				"port %s has %d fixed ips, a router interface requires exactly one",
				portID, len(found.FixedIPs)))
		}

		request = router.InterfaceRequest{PortID: portID}
	}

	log.Printf("[DEBUG] attaching VPC router %s interface", routerID)
	attached, err := router.AddInterface(ctx, client, routerID, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcRouterInterfaceV2Object, err))
	}

	d.SetId(attached.PortID)
	d.Set("port_id", attached.PortID)
	d.Set("network_id", attached.NetworkID)
	if attached.SubnetID != "" {
		d.Set("subnet_id", attached.SubnetID)
	}

	return resourceVPCRouterInterfaceV2Read(ctx, d, meta)
}

func resourceVPCRouterInterfaceV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting VPC router interface port %s", d.Id())
	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcRouterInterfaceV2Object)
	}

	d.Set("port_id", found.ID)
	d.Set("network_id", found.NetworkID)
	if found.DeviceID != "" {
		d.Set("router_id", found.DeviceID)
	}

	if subnetID := vpcRouterInterfaceV2SubnetFromPort(found.FixedIPs); subnetID != "" {
		d.Set("subnet_id", subnetID)
	}

	return nil
}

func resourceVPCRouterInterfaceV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	subnetID := d.Get("subnet_id").(string)

	// Clearing routes rewrites the whole list and detaching mutates the router,
	// so serialize updates per router.
	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	if d.Get("force_destroy").(bool) && subnetID != "" {
		if err := clearVPCRouterInterfaceV2Routes(ctx, client, routerID, subnetID); err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[DEBUG] detaching VPC router %s interface %s", routerID, d.Id())
	err = deleteVPCV2WithRetry(ctx, vpcRouterInterfaceV2Object, d.Id(),
		vpcV2StatusPollInterval, d.Timeout(schema.TimeoutDelete), vpcRouterInterfaceV2ConflictIsTransient,
		func(ctx context.Context) error {
			_, err := router.RemoveInterface(ctx, client, routerID, router.InterfaceRequest{PortID: d.Id()})

			return err
		})
	if err != nil {
		return diag.FromErr(errDeletingObject(vpcRouterInterfaceV2Object, d.Id(), err))
	}

	return nil
}

func vpcRouterInterfaceV2ConflictIsTransient(err error) bool {
	if !vpc.IsErrorClass(err, vpc.ErrorClassConflict) {
		return false
	}

	apiErr, ok := errors.AsType[*vpc.APIError](err)

	return ok && apiErr.Type == vpcRouterInterfaceV2InUseByFloatingIP
}

func resourceVPCRouterInterfaceV2ImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return nil, err
	}

	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		return nil, errGettingObject(vpcRouterInterfaceV2Object, d.Id(), err)
	}

	if found.DeviceID == "" {
		return nil, fmt.Errorf(
			"port %s is not attached to a router, so it is not a router interface", d.Id())
	}

	if len(found.FixedIPs) != 1 {
		return nil, fmt.Errorf(
			"port %s has %d fixed ips, so it carries more than one router interface "+
				"and cannot be imported as a single resource", d.Id(), len(found.FixedIPs))
	}

	d.Set("router_id", found.DeviceID)
	d.Set("port_id", found.ID)
	d.Set("network_id", found.NetworkID)
	d.Set("subnet_id", vpcRouterInterfaceV2SubnetFromPort(found.FixedIPs))

	return []*schema.ResourceData{d}, nil
}

func vpcRouterInterfaceV2SubnetFromPort(fixedIPs []port.FixedIP) string {
	if len(fixedIPs) != 1 {
		return ""
	}

	return fixedIPs[0].SubnetID
}

func clearVPCRouterInterfaceV2Routes(
	ctx context.Context, client *vpc.Client, routerID, subnetID string,
) error {
	attachedSubnet, err := subnet.Get(ctx, client, subnetID)
	if err != nil {
		return errGettingObject(vpcNetworkSubnetV2Object, subnetID, err)
	}

	_, cidr, err := net.ParseCIDR(attachedSubnet.CIDR)
	if err != nil {
		return fmt.Errorf("could not parse cidr %q of subnet %s: %w", attachedSubnet.CIDR, subnetID, err)
	}

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		return errGettingObject(vpcRouterV2Object, routerID, err)
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
		return errUpdatingObject(vpcRouterV2Object, routerID, err)
	}

	return nil
}
