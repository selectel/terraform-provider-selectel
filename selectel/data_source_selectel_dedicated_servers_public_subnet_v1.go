package selectel

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/dedicatedservers"
)

func dataSourceDedicatedServersPublicSubnetV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServersPublicSubnetV1Read,
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
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"subnet": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location_id": {
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
						"network_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"broadcast": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"gateway": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reserved_vrrp_ips": {
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

func dataSourceDedicatedServersPublicSubnetV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getDedicatedServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandDedicatedServersPublicSubnetsSearchFilter(d)

	subnets, _, err := dsClient.NetworkSubnets(ctx, filter.locationID)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectSubnet, err))
	}

	filteredSubnets, err := filterDedicatedServersPublicSubnets(subnets, filter)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error filtering subnets: %w", err))
	}

	subnetsFlatten := flattenDedicatedServersPublicSubnets(filteredSubnets)
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

type dedicatedServersPublicSubnetsSearchFilter struct {
	ip         string
	subnet     string
	locationID string
}

func expandDedicatedServersPublicSubnetsSearchFilter(d *schema.ResourceData) dedicatedServersPublicSubnetsSearchFilter {
	filter := dedicatedServersPublicSubnetsSearchFilter{}

	filterSet, ok := d.Get("filter").(*schema.Set)
	if !ok {
		return filter
	}

	if filterSet.Len() == 0 {
		return filter
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	ip, ok := resourceFilterMap["ip"]
	if ok {
		filter.ip = ip.(string)
	}

	subnet, ok := resourceFilterMap["subnet"]
	if ok {
		filter.subnet = subnet.(string)
	}

	locationID, ok := resourceFilterMap["location_id"]
	if ok {
		filter.locationID = locationID.(string)
	}

	return filter
}

func filterDedicatedServersPublicSubnets(subnets dedicatedservers.Subnets, filter dedicatedServersPublicSubnetsSearchFilter) (dedicatedservers.Subnets, error) {
	var filteredSubnets dedicatedservers.Subnets
	for _, subnet := range subnets {
		switch {
		case filter.ip == "" && filter.subnet == "":
			filteredSubnets = append(filteredSubnets, subnet)
			continue

		case filter.subnet != "" && filter.subnet == subnet.Subnet:
			filteredSubnets = append(filteredSubnets, subnet)
			continue
		}

		isIncluding, err := subnet.IsIncluding(filter.ip)
		if err != nil {
			return nil, fmt.Errorf("error checking if subnet %s includes IP %s: %w", subnet.UUID, filter.ip, err)
		}

		if isIncluding {
			filteredSubnets = append(filteredSubnets, subnet)
		}
	}

	return filteredSubnets, nil
}

func flattenDedicatedServersPublicSubnets(subnets dedicatedservers.Subnets) []interface{} {
	subnetsList := make([]interface{}, len(subnets))
	for i, subnet := range subnets {
		subnetsMap := make(map[string]interface{})
		subnetsMap["id"] = subnet.UUID
		subnetsMap["network_id"] = subnet.NetworkUUID
		subnetsMap["subnet"] = subnet.Subnet
		subnetsMap["broadcast"] = subnet.Broadcast.String()
		subnetsMap["gateway"] = subnet.Gateway.String()
		subnetsMap["reserved_vrrp_ips"] = subnet.ReservedVRRPIPAsStrings()

		subnetsList[i] = subnetsMap
	}

	return subnetsList
}
