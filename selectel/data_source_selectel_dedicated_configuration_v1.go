package selectel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
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
						"location_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsUUID,
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
						"config_name": {
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
	dsClient, diagErr := getDedicatedClient(d, meta, true)
	if diagErr != nil {
		return diagErr
	}

	filter, err := expandDedicatedConfigurationsSearchFilter(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Getting server configurations")

	serversList, _, err := dsClient.Servers(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"error getting list of servers configurations (without chips): %w", err,
		))
	}

	serverChipsList, _, err := dsClient.ServerChips(ctx)
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

type dedicatedConfigurationsFilter struct {
	deepFilter map[string]any
	name       string
	locationID string
}

func expandDedicatedConfigurationsSearchFilter(d *schema.ResourceData) (*dedicatedConfigurationsFilter, error) {
	filter := &dedicatedConfigurationsFilter{}

	filterRaw, ok := d.Get("deep_filter").(string)
	if ok && len(filterRaw) > 0 {
		filter.deepFilter = make(map[string]any)

		err := json.Unmarshal([]byte(filterRaw), &filter.deepFilter)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling deep_filter: %w", err)
		}
	}

	filterSet := d.Get("filter").(*schema.Set)
	if filterSet.Len() == 0 {
		return filter, nil
	}

	configurationFilterMap := filterSet.List()[0].(map[string]any)

	if v, ok := configurationFilterMap["name"].(string); ok {
		filter.name = v
	}

	if v, ok := configurationFilterMap["location_id"].(string); ok {
		filter.locationID = v
	}

	return filter, nil
}

func filterDedicatedConfigurations(list []dedicated.Server, filter *dedicatedConfigurationsFilter) []dedicated.Server {
	var filtered []dedicated.Server

	for _, entry := range list {
		match := true

		// Filter by name (partial match, case-insensitive)
		if filter.name != "" && !strings.Contains(strings.ToLower(entry.Name), strings.ToLower(filter.name)) {
			match = false
		}

		// Filter by location_id
		if match && filter.locationID != "" && !entry.IsLocationAvailable(filter.locationID) {
			match = false
		}

		// Filter by deep_filter
		if match && len(filter.deepFilter) > 0 {
			entryMap, err := reflect.StructToMap(entry)
			if err != nil || !reflect.IsSetContainsSubset(filter.deepFilter, entryMap) {
				match = false
			}
		}

		if match {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func flattenDedicatedConfigurations(list []dedicated.Server) []interface{} {
	res := make([]interface{}, len(list))
	for i, e := range list {
		sMap := make(map[string]interface{})
		sMap["id"] = e.ID
		sMap["name"] = e.Name
		sMap["config_name"] = e.ConfigName

		res[i] = sMap
	}

	return res
}
