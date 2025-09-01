package selectel

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/servers"
)

func dataSourceServersOSV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServersOSV1Read,
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
						"version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"configuration_id": {
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
			"os": {
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
						"arch": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"os": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"scripts_allowed": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"ssh_key_allowed": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"partitioning": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceServersOSV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandServersOperatingSystemsSearchFilter(d.Get("filter").(*schema.Set))

	log.Printf("[DEBUG] Getting %s '%#v'", objectOS, filter)

	opSystems, _, err := dsClient.OperatingSystems(ctx, servers.OperatingSystemsQuery{
		LocationID: filter.locationID,
		ServiceID:  filter.configurationID,
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(objectOS, err))
	}

	filteredOS, err := filterServersOperatingSystems(opSystems, filter)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error filtering locations: %w", err))
	}

	osFlatten := flattenServersOperatingSystems(filteredOS)
	if err := d.Set("os", osFlatten); err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(filteredOS))
	for _, e := range filteredOS {
		ids = append(ids, e.UUID)
	}

	checksum, err := stringListChecksum(ids)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(checksum)

	return nil
}

type serversOperatingSystemsFilter struct {
	name            string
	version         string
	configurationID string
	locationID      string
}

func expandServersOperatingSystemsSearchFilter(filterSet *schema.Set) serversOperatingSystemsFilter {
	filter := serversOperatingSystemsFilter{}
	if filterSet.Len() == 0 {
		return filter
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	name, ok := resourceFilterMap["name"]
	if ok {
		filter.name = name.(string)
	}

	configurationID, ok := resourceFilterMap["configuration_id"]
	if ok {
		filter.configurationID = configurationID.(string)
	}

	locationID, ok := resourceFilterMap["location_id"]
	if ok {
		filter.locationID = locationID.(string)
	}

	version, ok := resourceFilterMap["version"]
	if ok {
		filter.version = version.(string)
	}

	return filter
}

func filterServersOperatingSystems(list servers.OperatingSystems, filter serversOperatingSystemsFilter) (servers.OperatingSystems, error) {
	var filtered servers.OperatingSystems
	for _, entry := range list {
		if (filter.name == "" || entry.Name == filter.name) &&
			(filter.version == "" || entry.VersionValue == filter.version) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

func flattenServersOperatingSystems(list servers.OperatingSystems) []interface{} {
	res := make([]interface{}, len(list))
	for i, e := range list {
		sMap := make(map[string]interface{})
		sMap["id"] = e.UUID
		sMap["name"] = e.Name
		sMap["arch"] = e.Arch
		sMap["os"] = e.OSValue
		sMap["version"] = e.VersionValue
		sMap["scripts_allowed"] = e.ScriptAllowed
		sMap["ssh_key_allowed"] = e.IsSSHKeyAllowed
		sMap["partitioning"] = e.Partitioning

		res[i] = sMap
	}

	return res
}
