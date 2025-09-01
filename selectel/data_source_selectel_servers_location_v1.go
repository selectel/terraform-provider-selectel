package selectel

import (
	"context"
	"fmt"
	"log"

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

	filter := expandServersLocationSearchFilter(d.Get("filter").(*schema.Set))

	log.Print(msgGet(objectLocation, filter.name))

	locations, _, err := dsClient.Locations(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectLocation, err))
	}

	filteredLocations, err := filterServersLocations(locations, filter)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error filtering locations: %w", err))
	}

	locationsFlatten := flattenServersLocation(filteredLocations)
	if err := d.Set("locations", locationsFlatten); err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(locations))
	for _, e := range locations {
		ids = append(ids, e.UUID)
	}

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

func expandServersLocationSearchFilter(filterSet *schema.Set) serversLocationFilter {
	filter := serversLocationFilter{}
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

func filterServersLocations(list servers.Locations, filter serversLocationFilter) (servers.Locations, error) {
	var filtered servers.Locations
	for _, entry := range list {
		if filter.name == "" || entry.Name == filter.name {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
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
