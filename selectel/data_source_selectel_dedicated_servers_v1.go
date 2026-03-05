package selectel

import (
	"context"
	"fmt"
	"slices"
	"strings"

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
	dsClient, diagErr := getDedicatedClient(d, meta, true)
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

	reservedPrivateIPs, _, err := dsClient.NetworkReservedLocalIPs(ctx, "")
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
	if err = d.Set("servers", serversFlatten); err != nil {
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

func (f dedicatedServersSearchFilter) IsEmpty() bool {
	return f.name == "" &&
		f.ip == "" &&
		f.publicSubnet == "" &&
		f.privateSubnet == ""
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
	filteredServers := make([]dedicated.ResourceDetails, 0, len(servers))

	if filter.IsEmpty() {
		return servers, nil
	}

	for _, server := range servers {
		if !serverMatchesFilter(server, filter, reservedIPs) {
			continue
		}
		filteredServers = append(filteredServers, server)
	}

	return filteredServers, nil
}

func serverMatchesFilter(
	server dedicated.ResourceDetails, filter dedicatedServersSearchFilter, reservedIPs dedicated.ReservedIPs,
) bool {
	// Filter by name (partial match, case-insensitive)
	if filter.name != "" && !strings.Contains(strings.ToLower(server.Info), strings.ToLower(filter.name)) {
		return false
	}

	// Filter by IP or subnet
	if filter.ip != "" || filter.publicSubnet != "" || filter.privateSubnet != "" {
		return serverIPMatchesFilter(server, filter, reservedIPs)
	}

	return true
}

func serverIPMatchesFilter(
	server dedicated.ResourceDetails, filter dedicatedServersSearchFilter, reservedIPs dedicated.ReservedIPs,
) bool {
	for _, reserverIP := range reservedIPs {
		if server.UUID != reserverIP.ResourceUUID {
			continue
		}

		if filter.ip != "" && filter.ip == reserverIP.IP.String() {
			return true
		}
		if filter.publicSubnet != "" && filter.publicSubnet == reserverIP.Subnet {
			return true
		}
		if filter.privateSubnet != "" && filter.privateSubnet == reserverIP.Subnet {
			return true
		}
	}

	return false
}

func flattenDedicatedServers(
	servers []dedicated.ResourceDetails,
	reservedPublicIPs, reservedPrivateIPs dedicated.ReservedIPs,
) []any {
	serversList := make([]any, len(servers))

	for i, server := range servers {
		serverMap := make(map[string]any)
		serverMap["id"] = server.UUID
		serverMap["name"] = server.Info
		serverMap["configuration_id"] = server.ServiceUUID
		serverMap["location_id"] = server.LocationUUID

		publicIPs := make([]string, 0, len(reservedPublicIPs))
		for _, ip := range reservedPublicIPs {
			if ip.ResourceUUID == server.UUID {
				publicIPs = append(publicIPs, ip.IP.String())
			}
		}

		privateIPs := make([]string, 0, len(reservedPrivateIPs))
		for _, ip := range reservedPrivateIPs {
			if ip.ResourceUUID == server.UUID {
				privateIPs = append(privateIPs, ip.IP.String())
			}
		}

		serverMap["reserved_public_ips"] = publicIPs
		serverMap["reserved_private_ips"] = privateIPs

		serversList[i] = serverMap
	}

	return serversList
}
