package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
)

type availableExtensionSearchFilter struct {
	name string
}

func dataSourceDBaaSAvailableExtensionV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDBaaSAvailableExtensionV1Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
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
			"available_extensions": {
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
						"datastore_type_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"dependency_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceDBaaSAvailableExtensionV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	availableExtensions, err := dbaasClient.AvailableExtensions(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectAvailableExtensions, err))
	}

	availableExtensionIDs := []string{}
	for _, availableExtension := range availableExtensions {
		availableExtensionIDs = append(availableExtensionIDs, availableExtension.ID)
	}

	filter, err := expandAvailableExtensionSearchFilter(d.Get("filter").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	availableExtensions = filterAvailableExtensionByName(availableExtensions, filter.name)

	availableExtensionsFlatten := flattenAvailableExtensions(availableExtensions)
	if err := d.Set("available_extensions", availableExtensionsFlatten); err != nil {
		return diag.FromErr(err)
	}
	checksum, err := stringListChecksum(availableExtensionIDs)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}

func expandAvailableExtensionSearchFilter(filterSet *schema.Set) (availableExtensionSearchFilter, error) {
	filter := availableExtensionSearchFilter{}
	if filterSet.Len() == 0 {
		return filter, nil
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	name, ok := resourceFilterMap["name"]
	if ok {
		filter.name = name.(string)
	}

	return filter, nil
}

func filterAvailableExtensionByName(availableExtensions []dbaas.AvailableExtension, name string) []dbaas.AvailableExtension {
	if name == "" {
		return availableExtensions
	}

	var filteredAvailableExtensions []dbaas.AvailableExtension
	for _, ae := range availableExtensions {
		if ae.Name == name {
			filteredAvailableExtensions = append(filteredAvailableExtensions, ae)
		}
	}

	return filteredAvailableExtensions
}

func flattenAvailableExtensions(availableExtensions []dbaas.AvailableExtension) []interface{} {
	availableExtensionsList := make([]interface{}, len(availableExtensions))
	for i, availableExtension := range availableExtensions {
		availableExtensionsMap := make(map[string]interface{})
		availableExtensionsMap["id"] = availableExtension.ID
		availableExtensionsMap["name"] = availableExtension.Name
		availableExtensionsMap["datastore_type_ids"] = availableExtension.DatastoreTypeIDs
		availableExtensionsMap["dependency_ids"] = availableExtension.DependencyIDs

		availableExtensionsList[i] = availableExtensionsMap
	}

	return availableExtensionsList
}
