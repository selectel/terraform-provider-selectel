package selectel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/reflect"
)

func dataSourceDedicatedConfigurationV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedConfigurationV1Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"deep_filter": {
				Type:     schema.TypeString,
				Optional: true,
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

func dataSourceDedicatedConfigurationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getDedicatedClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter, err := expandDedicatedConfigurationsSearchFilter(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Getting server configurations")

	serversList, _, err := dsClient.ServersRaw(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"error getting list of servers configurations (without chips): %w", err,
		))
	}

	serverChipsList, _, err := dsClient.ServerChipsRaw(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"error getting list of servers configurations (with chips): %w", err))
	}

	serversList = append(serversList, serverChipsList...)

	filteredServers := filterDedicatedConfigurations(serversList, filter)

	serversFlatten := flattenDedicatedConfigurations(filteredServers)
	if err := d.Set("configurations", serversFlatten); err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(filteredServers))
	for _, e := range filteredServers {
		id, _ := e["uuid"].(string)
		ids = append(ids, id)
	}

	slices.Sort(ids)

	checksum, err := stringListChecksum(ids)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(checksum)

	return nil
}

type dedicatedConfigurationsFilter struct {
	deepFilter map[string]any
}

func expandDedicatedConfigurationsSearchFilter(d *schema.ResourceData) (*dedicatedConfigurationsFilter, error) {
	filter := &dedicatedConfigurationsFilter{
		deepFilter: make(map[string]any),
	}

	filterRaw, ok := d.Get("filter").(string)
	if !ok {
		return filter, nil
	}

	err := json.Unmarshal([]byte(filterRaw), &filter.deepFilter)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling deep_filter: %w", err)
	}

	return filter, nil
}

func filterDedicatedConfigurations(list []map[string]any, filter *dedicatedConfigurationsFilter) []map[string]any {
	var filteredRaw []map[string]any
	for _, entry := range list {
		if reflect.IsSetContainsSubset(filter.deepFilter, entry) {
			filteredRaw = append(filteredRaw, entry)
		}
	}

	return filteredRaw
}

func flattenDedicatedConfigurations(list []map[string]any) []interface{} {
	res := make([]interface{}, len(list))
	for i, e := range list {
		sMap := make(map[string]interface{})
		sMap["id"] = e["uuid"]
		sMap["name"] = e["name"]

		res[i] = sMap
	}

	return res
}
