package selectel

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/servers"
)

func dataSourceServersConfigurationV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServersConfigurationV1Read,
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
			"configurations": {
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
					},
				},
			},
		},
	}
}

func dataSourceServersConfigurationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandServersConfigurationSearchFilter(d)

	log.Print(msgGet(objectServer, filter.name))

	serversList, _, err := dsClient.Servers(ctx, false)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"error getting list of servers configurations (without chips): %w", err,
		))
	}

	serverChipsList, _, err := dsClient.Servers(ctx, true)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"error getting list of servers configurations (with chips): %w", err))
	}

	serversList = append(serversList, serverChipsList...)

	filteredServers := filterServersConfigurations(serversList, filter)

	serversFlatten := flattenServersConfiguration(filteredServers)
	if err := d.Set("configurations", serversFlatten); err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(filteredServers))
	for _, e := range filteredServers {
		ids = append(ids, e.ID)
	}

	slices.Sort(ids)

	checksum, err := stringListChecksum(ids)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(checksum)

	return nil
}

type serversConfigurationFilter struct {
	name string
}

func expandServersConfigurationSearchFilter(d *schema.ResourceData) serversConfigurationFilter {
	filter := serversConfigurationFilter{}

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

func filterServersConfigurations(list servers.Servers, filter serversConfigurationFilter) servers.Servers {
	var filtered servers.Servers
	for _, entry := range list {
		if filter.name == "" || entry.Name == filter.name {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func flattenServersConfiguration(list servers.Servers) []interface{} {
	res := make([]interface{}, len(list))
	for i, e := range list {
		sMap := make(map[string]interface{})
		sMap["id"] = e.ID
		sMap["name"] = e.Name

		res[i] = sMap
	}

	return res
}
