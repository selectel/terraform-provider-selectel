package selectel

import (
	"context"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/servers"
)

func dataSourceServersLocationV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServersLocationV1Read,
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

func dataSourceServersLocationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandServersLocationSearchFilter(d)

	log.Print(msgGet(objectLocation, filter.name))

	locations, _, err := dsClient.Locations(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectLocation, err))
	}

	filteredLocations := filterServersLocations(locations, filter)

	locationsFlatten := flattenServersLocation(filteredLocations)
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

type serversLocationFilter struct {
	name string
}

func expandServersLocationSearchFilter(d *schema.ResourceData) serversLocationFilter {
	filter := serversLocationFilter{}

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

func filterServersLocations(list servers.Locations, filter serversLocationFilter) servers.Locations {
	var filtered servers.Locations
	for _, entry := range list {
		if filter.name == "" || entry.Name == filter.name {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func flattenServersLocation(list servers.Locations) []interface{} {
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
