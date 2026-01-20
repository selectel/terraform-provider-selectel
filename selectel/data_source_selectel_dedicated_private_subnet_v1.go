package selectel

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dedicated "github.com/selectel/dedicated-go/pkg/v2"
)

func dataSourceDedicatedPrivateSubnetV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedPrivateSubnetV1Read,
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
						"location_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"subnet": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vlan": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			// computed
			"subnets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vlan": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reserved_ip": {
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

func dataSourceDedicatedPrivateSubnetV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dsClient, diagErr := getDedicatedClient(d, meta, true)
	if diagErr != nil {
		return diagErr
	}

	filter := expandDedicatedPrivateSubnetsSearchFilter(d)

	var allSubnets dedicated.Subnets

	nets, _, err := dsClient.Networks(ctx, filter.locationID, dedicated.NetworkTypeLocal, filter.vlan)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectNetwork, err))
	}

	// Get subnets for all networks in the location
	for _, network := range nets {
		localSubnets, _, err := dsClient.NetworkLocalSubnets(ctx, network.UUID)
		if err != nil {
			continue // Continue with other networks if one fails
		}
		allSubnets = append(allSubnets, localSubnets...)
	}

	filteredSubnets, err := filterDedicatedPrivateSubnets(allSubnets, filter)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error filtering subnets: %w", err))
	}

	subnetsFlatten, err := flattenDedicatedPrivateSubnets(ctx, filteredSubnets, dsClient)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("subnets", subnetsFlatten); err != nil {
		return diag.FromErr(err)
	}

	subnetsIDs := make([]string, 0, len(filteredSubnets))
	for _, subnet := range filteredSubnets {
		subnetsIDs = append(subnetsIDs, subnet.UUID)
	}

	slices.Sort(subnetsIDs)

	checksum, err := stringListChecksum(subnetsIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(checksum)

	return nil
}

type dedicatedPrivateSubnetsSearchFilter struct {
	locationID string
	ip         string
	subnet     string
	vlan       string
}

func expandDedicatedPrivateSubnetsSearchFilter(d *schema.ResourceData) dedicatedPrivateSubnetsSearchFilter {
	filter := dedicatedPrivateSubnetsSearchFilter{}

	filterSet, ok := d.Get("filter").(*schema.Set)
	if !ok {
		return filter
	}

	if filterSet.Len() == 0 {
		return filter
	}

	resourceFilterMap := filterSet.List()[0].(map[string]any)

	locationID, ok := resourceFilterMap["location_id"]
	if ok {
		filter.locationID = locationID.(string)
	}

	ip, ok := resourceFilterMap["ip"]
	if ok {
		filter.ip = ip.(string)
	}

	subnet, ok := resourceFilterMap["subnet"]
	if ok {
		filter.subnet = subnet.(string)
	}

	vlan, ok := resourceFilterMap["vlan"]
	if ok {
		filter.vlan = vlan.(string)
	}

	return filter
}

func filterDedicatedPrivateSubnets(subnets dedicated.Subnets, filter dedicatedPrivateSubnetsSearchFilter) (dedicated.Subnets, error) {
	var filteredSubnets dedicated.Subnets
	for _, subnet := range subnets {
		isIPIncluded := false
		if filter.ip != "" {
			var err error
			isIPIncluded, err = subnet.IsIncluding(filter.ip)
			if err != nil {
				return nil, fmt.Errorf("error checking if subnet %s includes IP %s: %w", subnet.UUID, filter.ip, err)
			}
		}

		if (filter.subnet == "" || filter.subnet == subnet.Subnet) && (filter.ip == "" || isIPIncluded) {
			filteredSubnets = append(filteredSubnets, subnet)

			continue
		}
	}

	return filteredSubnets, nil
}

func flattenDedicatedPrivateSubnets(ctx context.Context, subnets dedicated.Subnets, dsClient *dedicated.ServiceClient) ([]map[string]any, error) {
	subnetsList := make([]map[string]any, 0, len(subnets))

	for _, subnet := range subnets {
		subnetMap := make(map[string]any)

		subnetMap["id"] = subnet.UUID
		subnetMap["vlan"] = strconv.Itoa(subnet.Network)
		subnetMap["subnet"] = subnet.Subnet

		reservedIPs, _, err := dsClient.NetworkSubnetLocalReservedIPs(ctx, subnet.UUID)
		if err != nil {
			return nil, err
		}

		var ips []string
		for _, reservedIP := range reservedIPs {
			ips = append(ips, reservedIP.IP.String())
		}

		subnetMap["reserved_ip"] = ips

		subnetsList = append(subnetsList, subnetMap)
	}

	return subnetsList, nil
}
