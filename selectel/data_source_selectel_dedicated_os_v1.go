package selectel

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dedicated "github.com/selectel/dedicated-go/pkg/v2"
)

func dataSourceDedicatedOSV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedOSV1Read,
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
						"version_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"version_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"version_name_regex": {
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
						"version_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version_name": {
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

func dataSourceDedicatedOSV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getDedicatedClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandDedicatedOperatingSystemsSearchFilter(d)

	log.Printf("[DEBUG] Getting %s '%#v'", objectOS, filter)

	opSystems, _, err := dsClient.OperatingSystems(ctx, &dedicated.OperatingSystemsQuery{
		LocationID: filter.locationID,
		ServiceID:  filter.configurationID,
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(objectOS, err))
	}

	filteredOS, err := filterDedicatedOperatingSystems(opSystems, filter)
	if err != nil {
		return diag.FromErr(err)
	}

	osFlatten := flattenDedicatedOperatingSystems(filteredOS)
	if err := d.Set("os", osFlatten); err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(filteredOS))
	for _, e := range filteredOS {
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

type dedicatedOperatingSystemsFilter struct {
	name             string
	versionValue     string
	versionName      string
	versionNameRegex string
	configurationID  string
	locationID       string
}

func expandDedicatedOperatingSystemsSearchFilter(d *schema.ResourceData) dedicatedOperatingSystemsFilter {
	filter := dedicatedOperatingSystemsFilter{}

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

	configurationID, ok := resourceFilterMap["configuration_id"]
	if ok {
		filter.configurationID = configurationID.(string)
	}

	locationID, ok := resourceFilterMap["location_id"]
	if ok {
		filter.locationID = locationID.(string)
	}

	versionValue, ok := resourceFilterMap["version_value"]
	if ok {
		filter.versionValue = versionValue.(string)
	}

	versionName, ok := resourceFilterMap["version_name"]
	if ok {
		filter.versionName = versionName.(string)
	}

	versionNameRegex, ok := resourceFilterMap["version_name_regex"]
	if ok {
		filter.versionNameRegex = versionNameRegex.(string)
	}

	return filter
}

func filterDedicatedOperatingSystems(list dedicated.OperatingSystems, filter dedicatedOperatingSystemsFilter) (dedicated.OperatingSystems, error) {
	var filtered dedicated.OperatingSystems
	for _, entry := range list {
		isVersionNameRegexMatches := true
		if filter.versionNameRegex != "" {
			matches, err := regexp.MatchString(filter.versionNameRegex, entry.VersionName)
			if err != nil {
				return nil, fmt.Errorf("error parsing version name regex: %s", err)
			}

			isVersionNameRegexMatches = matches
		}

		isVersionNameMatches := filter.versionName == "" ||
			strings.Contains(strings.ToLower(entry.VersionName), strings.ToLower(filter.versionName))

		isNameMatches := filter.name == "" || entry.Name == filter.name

		isVersionValueMatches := filter.versionValue == "" || entry.VersionValue == filter.versionValue

		if isNameMatches && isVersionNameMatches && isVersionNameRegexMatches && isVersionValueMatches {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

func flattenDedicatedOperatingSystems(list dedicated.OperatingSystems) []interface{} {
	res := make([]interface{}, len(list))
	for i, e := range list {
		sMap := make(map[string]interface{})
		sMap["id"] = e.UUID
		sMap["name"] = e.Name
		sMap["arch"] = e.Arch
		sMap["os"] = e.OSValue
		sMap["version_value"] = e.VersionValue
		sMap["version_name"] = e.VersionName
		sMap["scripts_allowed"] = e.ScriptAllowed
		sMap["ssh_key_allowed"] = e.IsSSHKeyAllowed
		sMap["partitioning"] = e.Partitioning

		res[i] = sMap
	}

	return res
}
