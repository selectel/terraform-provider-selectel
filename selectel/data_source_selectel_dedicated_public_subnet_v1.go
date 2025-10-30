package selectel

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dedicated "github.com/selectel/dedicated-go/pkg/v2"
)

func dataSourceDedicatedPublicSubnetV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedPublicSubnetV1Read,
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
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDedicatedPublicSubnetV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getDedicatedClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandDedicatedPublicSubnetsSearchFilter(d)

	subnets, _, err := dsClient.NetworkSubnets(ctx, filter.locationID)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectSubnet, err))
	}

	filteredSubnets, err := filterDedicatedPublicSubnets(subnets, filter)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error filtering subnets: %w", err))
	}

	subnetsFlatten := flattenDedicatedPublicSubnets(filteredSubnets, filter)
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

type dedicatedPublicSubnetsSearchFilter struct {
	ip         string
	subnet     string
	locationID string
}

func expandDedicatedPublicSubnetsSearchFilter(d *schema.ResourceData) dedicatedPublicSubnetsSearchFilter {
	filter := dedicatedPublicSubnetsSearchFilter{}

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

func filterDedicatedPublicSubnets(subnets dedicated.Subnets, filter dedicatedPublicSubnetsSearchFilter) (dedicated.Subnets, error) {
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

		if (filter.subnet == "" || filter.subnet == subnet.Subnet) &&
			(filter.ip == "" || isIPIncluded) {
			filteredSubnets = append(filteredSubnets, subnet)

			continue
		}
	}

	return filteredSubnets, nil
}

func flattenDedicatedPublicSubnets(subnets dedicated.Subnets, filter dedicatedPublicSubnetsSearchFilter) []interface{} {
	subnetsList := make([]interface{}, len(subnets))
	for i, subnet := range subnets {
		subnetMap := make(map[string]interface{})
		subnetMap["id"] = subnet.UUID
		subnetMap["network_id"] = subnet.NetworkUUID
		subnetMap["subnet"] = subnet.Subnet
		subnetMap["broadcast"] = subnet.Broadcast.String()
		subnetMap["gateway"] = subnet.Gateway.String()
		subnetMap["reserved_vrrp_ips"] = subnet.ReservedVRRPIPAsStrings()

		if filter.ip != "" {
			subnetMap["ip"] = filter.ip
		}

		subnetsList[i] = subnetMap
	}

	return subnetsList
}
