package selectel

import (
	"context"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dedicated "github.com/selectel/dedicated-go/pkg/v2"
)

func dataSourceDedicatedLocationV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedLocationV1Read,
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
					},
				},
			},
			// computed
			"locations": {
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
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"visibility": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDedicatedLocationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getDedicatedClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandDedicatedLocationsSearchFilter(d)

	log.Print(msgGet(objectLocation, filter.name))

	locations, _, err := dsClient.Locations(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectLocation, err))
	}

	filteredLocations := filterDedicatedLocations(locations, filter)

	locationsFlatten := flattenDedicatedLocations(filteredLocations)
	if err := d.Set("locations", locationsFlatten); err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(filteredLocations))
	for _, e := range filteredLocations {
		ids = append(ids, e.UUID)
	}

	slices.Sort(ids)

	checksum, err := stringListChecksum(ids)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(checksum)

	return nil
}

type dedicatedLocationsFilter struct {
	name string
}

func expandDedicatedLocationsSearchFilter(d *schema.ResourceData) dedicatedLocationsFilter {
	filter := dedicatedLocationsFilter{}

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

	return filter
}

func filterDedicatedLocations(list dedicated.Locations, filter dedicatedLocationsFilter) dedicated.Locations {
	var filtered dedicated.Locations
	for _, entry := range list {
		if filter.name == "" || entry.Name == filter.name {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func flattenDedicatedLocations(list dedicated.Locations) []interface{} {
	res := make([]interface{}, len(list))
	for i, e := range list {
		sMap := make(map[string]interface{})
		sMap["id"] = e.UUID
		sMap["name"] = e.Name
		sMap["description"] = e.Description
		sMap["visibility"] = e.Visibility

		res[i] = sMap
	}

	return res
}
