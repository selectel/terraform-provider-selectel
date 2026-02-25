package selectel

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
)

func dataSourceDedicatedServersV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServersV1Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"configuration_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"public_subnet": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"private_subnet": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			// computed
			"servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"configuration_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"location_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reserved_public_ips": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"reserved_private_ips": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceDedicatedServersV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dsClient, diagErr := getDedicatedClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandDedicatedServersSearchFilter(d)

	servers, _, err := dsClient.ResourcesList(ctx, filter.locationID, filter.configurationID)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectDedicatedServer, err))
	}

	reservedIPs := make(dedicated.ReservedIPs, 0)

	reservedPublicIPs, _, err := dsClient.NetworkReservedIPs(ctx, filter.locationID, "")
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting reserved public IPs: %w", err))
	}
	if len(reservedPublicIPs) > 0 {
		reservedIPs = append(reservedIPs, reservedPublicIPs...)
	}

	reservedPrivateIPs, _, err := dsClient.NetworkReservedLocalIPs(ctx, filter.locationID, "")
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting reserved private IPs: %w", err))
	}
	if len(reservedPrivateIPs) > 0 {
		reservedIPs = append(reservedIPs, reservedPrivateIPs...)
	}

	filteredServers, err := filterDedicatedServers(servers, filter, reservedIPs)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error filtering servers: %w", err))
	}

	serversFlatten := flattenDedicatedServers(filteredServers, reservedPublicIPs, reservedPrivateIPs)
	if err := d.Set("servers", serversFlatten); err != nil {
		return diag.FromErr(err)
	}

	serversIDs := make([]string, 0, len(filteredServers))
	for _, server := range filteredServers {
		serversIDs = append(serversIDs, server.UUID)
	}

	slices.Sort(serversIDs)

	checksum, err := stringListChecksum(serversIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(checksum)

	return nil
}

type dedicatedServersSearchFilter struct {
	name            string
	ip              string
	locationID      string
	configurationID string
	publicSubnet    string
	privateSubnet   string
}

func expandDedicatedServersSearchFilter(d *schema.ResourceData) dedicatedServersSearchFilter {
	filter := dedicatedServersSearchFilter{}

	filterSet, ok := d.Get("filter").(*schema.Set)
	if !ok {
		return filter
	}

	if filterSet.Len() == 0 {
		return filter
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	name, ok := resourceFilterMap["name"]
	if ok {
		filter.name = name.(string)
	}

	ip, ok := resourceFilterMap["ip"]
	if ok {
		filter.ip = ip.(string)
	}

	locationID, ok := resourceFilterMap["location_id"]
	if ok {
		filter.locationID = locationID.(string)
	}

	configurationID, ok := resourceFilterMap["configuration_id"]
	if ok {
		filter.configurationID = configurationID.(string)
	}

	publicSubnet, ok := resourceFilterMap["public_subnet"]
	if ok {
		filter.publicSubnet = publicSubnet.(string)
	}

	privateSubnet, ok := resourceFilterMap["private_subnet"]
	if ok {
		filter.privateSubnet = privateSubnet.(string)
	}

	return filter
}

func filterDedicatedServers(
	servers []dedicated.ResourceDetails, filter dedicatedServersSearchFilter, reservedIPs dedicated.ReservedIPs,
) ([]dedicated.ResourceDetails, error) {
	var filteredServers []dedicated.ResourceDetails
	for _, server := range servers {
		if filter.name != "" && filter.name == server.Info {
			filteredServers = append(filteredServers, server)
		}

		if filter.ip != "" {
			for _, reserverIP := range reservedIPs {
				if filter.ip == reserverIP.IP.String() ||
					filter.privateSubnet == reserverIP.Subnet ||
					filter.publicSubnet == reserverIP.Subnet {
					filteredServers = append(filteredServers, server)
				}
			}
		}
	}

	return filteredServers, nil
}

func flattenDedicatedServers(
	servers []dedicated.ResourceDetails, reservedPublicIPs, reservedPrivateIPs dedicated.ReservedIPs,
) []any {
	serversList := make([]interface{}, len(servers))
	reservedPublicIPsStrings := make([]string, len(reservedPublicIPs))
	reservedPrivateIPsStrings := make([]string, len(reservedPrivateIPs))

	for i, server := range servers {
		serverMap := make(map[string]interface{})
		serverMap["id"] = server.UUID
		serverMap["name"] = server.Info
		serverMap["configuration_id"] = server.ServiceUUID
		serverMap["location_id"] = server.LocationUUID

		if len(reservedPublicIPs) > 0 {
			for _, ip := range reservedPublicIPs {
				reservedPublicIPsStrings = append(reservedPublicIPsStrings, ip.IP.String())
			}
		}

		if len(reservedPrivateIPs) > 0 {
			for _, ip := range reservedPrivateIPs {
				reservedPrivateIPsStrings = append(reservedPrivateIPsStrings, ip.IP.String())
			}
		}

		serverMap["reserved_public_ips"] = reservedPublicIPsStrings
		serverMap["reserved_private_ips"] = reservedPrivateIPsStrings

		serversList[i] = serverMap
	}

	return serversList
}
