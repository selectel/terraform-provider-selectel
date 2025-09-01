package selectel

import (
	"context"
	"fmt"
	"log"

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
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"is_server_chip": {
							Type:     schema.TypeBool,
							Required: true,
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

	filter := expandServersConfigurationSearchFilter(d.Get("filter").(*schema.Set))

	objectServerName := objectServerChip
	if !filter.isServerChip {
		objectServerName = objectServer
	}

	log.Print(msgGet(objectServerName, filter.name))

	serversList, _, err := dsClient.Servers(ctx, filter.isServerChip)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectServerName, err))
	}

	filteredServers, err := filterServersConfigurations(serversList, filter)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error filtering servers: %w", err))
	}

	serversFlatten := flattenServersConfiguration(filteredServers)
	if err := d.Set("configurations", serversFlatten); err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(serversList))
	for _, e := range serversList {
		ids = append(ids, e.ID)
	}

	checksum, err := stringListChecksum(ids)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(checksum)

	return nil
}

type serversConfigurationFilter struct {
	name         string
	isServerChip bool
}

func expandServersConfigurationSearchFilter(filterSet *schema.Set) serversConfigurationFilter {
	filter := serversConfigurationFilter{}
	if filterSet.Len() == 0 {
		return filter
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	name, ok := resourceFilterMap["name"]
	if ok {
		filter.name = name.(string)
	}

	isServerChip, ok := resourceFilterMap["is_server_chip"]
	if ok {
		filter.isServerChip = isServerChip.(bool)
	}

	return filter
}

func filterServersConfigurations(list servers.Servers, filter serversConfigurationFilter) (servers.Servers, error) {
	var filtered servers.Servers
	for _, entry := range list {
		if filter.name == "" || entry.Name == filter.name {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
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
