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
						"configuration": {
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
						"configuration": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"location": {
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

	servers, _, err := dsClient.ResourcesList(ctx, filter.locationID, "")
	if err != nil {
		return diag.FromErr(errGettingObjects(objectDedicatedServer, err))
	}

	serverConfigs, _, err := dsClient.Servers(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting server configurations: %w", err))
	}

	serverChipConfigs, _, err := dsClient.ServerChips(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting server chip configurations: %w", err))
	}

	configNameByUUID := buildConfigNameByUUID(serverConfigs, serverChipConfigs)

	locations, _, err := dsClient.Locations(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting locations: %w", err))
	}

	locationNameByUUID := make(map[string]string, len(locations))
	for _, loc := range locations {
		locationNameByUUID[loc.UUID] = loc.Name
	}

	reservedPublicIPs, _, err := dsClient.NetworkReservedIPs(ctx, filter.locationID, "")
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting reserved public IPs: %w", err))
	}

	reservedPrivateIPs, _, err := dsClient.NetworkReservedLocalIPs(ctx, "")
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting reserved private IPs: %w", err))
	}

	filteredServers, err := filterDedicatedServers(servers, filter, reservedPublicIPs, reservedPrivateIPs, configNameByUUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error filtering servers: %w", err))
	}

	serversFlatten := flattenDedicatedServers(filteredServers, reservedPublicIPs, reservedPrivateIPs, configNameByUUID, locationNameByUUID)
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
	name          string
	ip            string
	locationID    string
	configuration string
	publicSubnet  string
	privateSubnet string
}

// IsEmpty reports whether all in-memory filter fields are unset.
// locationID is excluded because it is applied at the API call layer, not in-memory.
func (f dedicatedServersSearchFilter) IsEmpty() bool {
	return f.name == "" &&
		f.ip == "" &&
		f.configuration == "" &&
		f.publicSubnet == "" &&
		f.privateSubnet == ""
}

func expandDedicatedServersSearchFilter(d *schema.ResourceData) dedicatedServersSearchFilter {
	filter := dedicatedServersSearchFilter{}

	filterSet, ok := d.Get("filter").(*schema.Set)
	if !ok || filterSet.Len() == 0 {
		return filter
	}

	m := filterSet.List()[0].(map[string]interface{})

	if v, ok := m["name"]; ok {
		filter.name = v.(string)
	}
	if v, ok := m["ip"]; ok {
		filter.ip = v.(string)
	}
	if v, ok := m["location_id"]; ok {
		filter.locationID = v.(string)
	}
	if v, ok := m["configuration"]; ok {
		filter.configuration = v.(string)
	}
	if v, ok := m["public_subnet"]; ok {
		filter.publicSubnet = v.(string)
	}
	if v, ok := m["private_subnet"]; ok {
		filter.privateSubnet = v.(string)
	}

	return filter
}

func filterDedicatedServers(
	servers []dedicated.ResourceDetails, filter dedicatedServersSearchFilter,
	publicReservedIPs, privateReservedIPs dedicated.ReservedIPs,
	configNameByUUID map[string]string,
) ([]dedicated.ResourceDetails, error) {
	if filter.IsEmpty() {
		return servers, nil
	}

	filteredServers := make([]dedicated.ResourceDetails, 0, len(servers))
	for _, server := range servers {
		if serverMatchesFilter(server, filter, publicReservedIPs, privateReservedIPs, configNameByUUID) {
			filteredServers = append(filteredServers, server)
		}
	}

	return filteredServers, nil
}

func serverMatchesFilter(
	server dedicated.ResourceDetails, filter dedicatedServersSearchFilter,
	publicReservedIPs, privateReservedIPs dedicated.ReservedIPs,
	configNameByUUID map[string]string,
) bool {
	if filter.name != "" {
		name := server.UserDesc
		if name == "" {
			name = server.Info
		}
		if !strings.Contains(strings.ToLower(name), strings.ToLower(filter.name)) {
			return false
		}
	}

	if filter.configuration != "" {
		configName, ok := configNameByUUID[server.ServiceUUID]
		if !ok {
			return false
		}
		if !strings.Contains(strings.ToLower(configName), strings.ToLower(filter.configuration)) {
			return false
		}
	}

	if filter.ip != "" || filter.publicSubnet != "" || filter.privateSubnet != "" {
		return serverIPMatchesFilter(server, filter, publicReservedIPs, privateReservedIPs)
	}

	return true
}

func serverIPMatchesFilter(
	server dedicated.ResourceDetails, filter dedicatedServersSearchFilter,
	publicReservedIPs, privateReservedIPs dedicated.ReservedIPs,
) bool {
	for _, rip := range publicReservedIPs {
		if server.UUID != rip.ResourceUUID {
			continue
		}
		if filter.ip != "" && filter.ip == rip.IP.String() {
			return true
		}
		if filter.publicSubnet != "" && filter.publicSubnet == rip.Subnet {
			return true
		}
	}

	for _, rip := range privateReservedIPs {
		if server.UUID != rip.ResourceUUID {
			continue
		}
		if filter.ip != "" && filter.ip == rip.IP.String() {
			return true
		}
		if filter.privateSubnet != "" && filter.privateSubnet == rip.Subnet {
			return true
		}
	}

	return false
}

func flattenDedicatedServers(
	servers []dedicated.ResourceDetails,
	reservedPublicIPs, reservedPrivateIPs dedicated.ReservedIPs,
	configNameByUUID, locationNameByUUID map[string]string,
) []any {
	serversList := make([]any, len(servers))

	for i, server := range servers {
		serverMap := make(map[string]any)
		serverMap["id"] = server.UUID

		name := server.UserDesc
		if name == "" {
			name = server.Info
		}
		serverMap["name"] = name

		serverMap["configuration_id"] = server.ServiceUUID
		serverMap["location_id"] = server.LocationUUID
		serverMap["configuration"] = configNameByUUID[server.ServiceUUID]
		serverMap["location"] = locationNameByUUID[server.LocationUUID]

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

func buildConfigNameByUUID(servers, serverChips []dedicated.Server) map[string]string {
	m := make(map[string]string, len(servers)+len(serverChips))
	for _, s := range servers {
		m[s.ID] = s.Name
	}
	// serverChips entries overwrite servers entries on UUID collision (service types are disjoint in practice).
	for _, s := range serverChips {
		m[s.ID] = s.Name
	}

	return m
}
