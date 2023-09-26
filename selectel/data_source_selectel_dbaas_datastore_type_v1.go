package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
)

type datastoreTypeSearchFilter struct {
	engine  string
	version string
}

func dataSourceDBaaSDatastoreTypeV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDBaaSDatastoreTypeV1Read,
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
						"engine": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"datastore_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDBaaSDatastoreTypeV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	datastoreTypes, err := dbaasClient.DatastoreTypes(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectDatastoreTypes, err))
	}

	datastoreTypeIDs := []string{}
	for _, datastoreType := range datastoreTypes {
		datastoreTypeIDs = append(datastoreTypeIDs, datastoreType.ID)
	}

	filter, err := expandDatastoreTypeSearchFilter(d.Get("filter").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	datastoreTypes = filterDatastoreTypesByEngine(datastoreTypes, filter.engine)
	datastoreTypes = filterDatastoreTypesByVersion(datastoreTypes, filter.version)

	datastoreTypesFlatten := flattenDBaaSDatastoreTypes(datastoreTypes)
	if err := d.Set("datastore_types", datastoreTypesFlatten); err != nil {
		return diag.FromErr(err)
	}
	checksum, err := stringListChecksum(datastoreTypeIDs)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}

func expandDatastoreTypeSearchFilter(filterSet *schema.Set) (datastoreTypeSearchFilter, error) {
	filter := datastoreTypeSearchFilter{}
	if filterSet.Len() == 0 {
		return filter, nil
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	engine, ok := resourceFilterMap["engine"]
	if ok {
		filter.engine = engine.(string)
	}

	version, ok := resourceFilterMap["version"]
	if ok {
		filter.version = version.(string)
	}

	return filter, nil
}

func filterDatastoreTypesByVersion(datastoreTypes []dbaas.DatastoreType, version string) []dbaas.DatastoreType {
	if version == "" {
		return datastoreTypes
	}

	var filteredDatastoreTypes []dbaas.DatastoreType
	for _, dt := range datastoreTypes {
		if dt.Version == version {
			filteredDatastoreTypes = append(filteredDatastoreTypes, dt)
		}
	}

	return filteredDatastoreTypes
}

func filterDatastoreTypesByEngine(datastoreTypes []dbaas.DatastoreType, engine string) []dbaas.DatastoreType {
	if engine == "" {
		return datastoreTypes
	}

	var filteredDatastoreTypes []dbaas.DatastoreType
	for _, dt := range datastoreTypes {
		if dt.Engine == engine {
			filteredDatastoreTypes = append(filteredDatastoreTypes, dt)
		}
	}

	return filteredDatastoreTypes
}

func flattenDBaaSDatastoreTypes(datastoreTypes []dbaas.DatastoreType) []interface{} {
	datastoreTypesList := make([]interface{}, len(datastoreTypes))
	for i, datastoreType := range datastoreTypes {
		datastoreTypesMap := make(map[string]interface{})
		datastoreTypesMap["id"] = datastoreType.ID
		datastoreTypesMap["engine"] = datastoreType.Engine
		datastoreTypesMap["version"] = datastoreType.Version

		datastoreTypesList[i] = datastoreTypesMap
	}

	return datastoreTypesList
}
