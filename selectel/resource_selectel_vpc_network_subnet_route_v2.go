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
	"github.com/selectel/vpc-go/pkg/v2/subnet"
)

const (
	vpcNetworkSubnetRouteV2Object      = "VPC network subnet route"
	vpcNetworkSubnetRouteV2IDSeparator = "-route-"
)

var vpcNetworkSubnetRouteDocs = resourceDocs{Name: "route"}

func resourceVPCNetworkSubnetRouteV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCNetworkSubnetRouteV2Create,
		ReadContext:   resourceVPCNetworkSubnetRouteV2Read,
		DeleteContext: resourceVPCNetworkSubnetRouteV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCNetworkSubnetRouteV2ImportState,
		},
		Description: "Creates and manages a single route of a subnet using public API v2. " +
			"For more information about static routes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/static-routes/).\n\n" +
			"The resource manages a single entry of the subnet's `host_routes` list and leaves the other entries alone. " +
			"Do not manage the same routes both with this resource and the `host_routes` argument of `selectel_vpc_network_subnet_v2`.\n\n" +
			upgradeGuideNote,
		Schema: vpcNetworkSubnetRouteDocs.withDocsHints(map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Description: vpcNetworkSubnetRouteDocs.idDescription() + ". " +
					"Composite identifier of the subnet route association in format `<subnet_id>-route-<destination_cidr>-<next_hop>`, for example, `8f6c1e9b-2a3d-4f50-9b8a-1c2d3e4f5a6b-route-10.0.0.0/8-192.168.199.10`.",
			},
			"project_id": projectIDResourceSchema(),
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Pool where the subnet is located, for example, `" + exampleRegion + "`. " + regionLearnMore,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "Unique identifier of the subnet the route belongs to. " +
					"Retrieved from the [selectel_vpc_network_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) resource.",
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
				Description:  "IP address of the next hop gateway, for example, `192.168.199.10`.",
			},
		}),
	}
}

func resourceVPCNetworkSubnetRouteV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	subnetID := d.Get("subnet_id").(string)
	route := subnet.HostRoute{
		Destination: d.Get("destination_cidr").(string),
		NextHop:     d.Get("next_hop").(string),
	}

	// Serialize read-modify-write updates per subnet.
	selMutexKV.Lock(subnetID)
	defer selMutexKV.Unlock(subnetID)

	found, err := subnet.Get(ctx, client, subnetID)
	if err != nil {
		return diag.FromErr(errGettingObject(vpcNetworkSubnetV2Object, subnetID, err))
	}

	if slices.Contains(found.HostRoutes, route) {
		return diag.FromErr(fmt.Errorf(
			"%s with destination_cidr '%s' and next_hop '%s' already exists on subnet '%s'",
			vpcNetworkSubnetRouteV2Object, route.Destination, route.NextHop, subnetID))
	}

	routes := append(append([]subnet.HostRoute{}, found.HostRoutes...), route)

	log.Printf("[DEBUG] adding route to VPC network subnet %s", subnetID)
	if _, err := subnet.Update(ctx, client, subnetID, subnet.UpdateRequest{
		HostRoutes: &routes,
	}); err != nil {
		return diag.FromErr(errCreatingObject(vpcNetworkSubnetRouteV2Object, err))
	}

	d.SetId(vpcNetworkSubnetRouteV2ID(subnetID, route))

	return resourceVPCNetworkSubnetRouteV2Read(ctx, d, meta)
}

func resourceVPCNetworkSubnetRouteV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
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
		return vpcCheckDeleted(d, err, vpcNetworkSubnetV2Object)
	}

	if !slices.Contains(found.HostRoutes, route) {
		d.SetId("")

		return nil
	}

	d.Set("subnet_id", subnetID)
	d.Set("destination_cidr", route.Destination)
	d.Set("next_hop", route.NextHop)

	return nil
}

func resourceVPCNetworkSubnetRouteV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
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
		if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return nil
		}

		return diag.FromErr(errGettingObject(vpcNetworkSubnetV2Object, subnetID, err))
	}

	index := slices.Index(found.HostRoutes, route)
	if index < 0 {
		return nil
	}

	routes := make([]subnet.HostRoute, 0, len(found.HostRoutes)-1)
	routes = append(routes, found.HostRoutes[:index]...)
	routes = append(routes, found.HostRoutes[index+1:]...)

	log.Printf("[DEBUG] removing route from VPC network subnet %s", subnetID)
	if _, err := subnet.Update(ctx, client, subnetID, subnet.UpdateRequest{
		HostRoutes: &routes,
	}); err != nil {
		return diag.FromErr(errDeletingObject(vpcNetworkSubnetRouteV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCNetworkSubnetRouteV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	subnetID, route, err := parseVPCNetworkSubnetRouteV2ID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("subnet_id", subnetID)
	d.Set("destination_cidr", route.Destination)
	d.Set("next_hop", route.NextHop)

	return []*schema.ResourceData{d}, nil
}

func vpcNetworkSubnetRouteV2ID(subnetID string, route subnet.HostRoute) string {
	return fmt.Sprintf("%s%s%s-%s",
		subnetID, vpcNetworkSubnetRouteV2IDSeparator, route.Destination, route.NextHop)
}

func parseVPCNetworkSubnetRouteV2ID(id string) (string, subnet.HostRoute, error) {
	subnetID, remainder, found := strings.Cut(id, vpcNetworkSubnetRouteV2IDSeparator)
	if !found || subnetID == "" {
		return "", subnet.HostRoute{}, errParseVPCNetworkSubnetRouteV2ID(id)
	}

	parts := strings.Split(remainder, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", subnet.HostRoute{}, errParseVPCNetworkSubnetRouteV2ID(id)
	}

	return subnetID, subnet.HostRoute{Destination: parts[0], NextHop: parts[1]}, nil
}

func errParseVPCNetworkSubnetRouteV2ID(id string) error {
	return fmt.Errorf(
		"unable to parse %s ID '%s': expected the format <subnet_id>%s<destination_cidr>-<next_hop>",
		vpcNetworkSubnetRouteV2Object, id, vpcNetworkSubnetRouteV2IDSeparator)
}
