package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dbaas_v2 "github.com/selectel/dbaas-go/v2/common"
)

type dbaasV2DatastoreTypeSearchFilter struct {
	engine  string
	version string
}

func dataSourceDBaaSV2DatastoreType() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDBaaSV2DatastoreTypeRead,
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

func dataSourceDBaaSV2DatastoreTypeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	response, err := dbaasClient.DatastoreType.GetDatastoreTypeList(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectDatastoreTypes, err))
	}

	if response.Errors != "" {
		log.Printf("[WARN] datastore types got with error: %s", response.Errors)
	}

	datastoreTypes := response.DatastoreTypes

	datastoreTypeIDs := []string{}
	for _, datastoreType := range datastoreTypes {
		datastoreTypeIDs = append(datastoreTypeIDs, datastoreType.ID)
	}

	filter, err := expandDBaaSV2DatastoreTypeSearchFilter(d.Get("filter").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	datastoreTypes = filterDBaaSV2DatastoreTypesByEngine(datastoreTypes, filter.engine)
	datastoreTypes = filterDBaaSV2DatastoreTypesByVersion(datastoreTypes, filter.version)

	datastoreTypesFlatten := flattenDBaaSV2DatastoreTypes(datastoreTypes)
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

func filterDBaaSV2DatastoreTypesByVersion(datastoreTypes []dbaas_v2.DatastoreTypeResponse, version string) []dbaas_v2.DatastoreTypeResponse {
	if version == "" {
		return datastoreTypes
	}

	var filteredDatastoreTypes []dbaas_v2.DatastoreTypeResponse
	for _, dt := range datastoreTypes {
		if dt.Version == version {
			filteredDatastoreTypes = append(filteredDatastoreTypes, dt)
		}
	}

	return filteredDatastoreTypes
}

func filterDBaaSV2DatastoreTypesByEngine(datastoreTypes []dbaas_v2.DatastoreTypeResponse, engine string) []dbaas_v2.DatastoreTypeResponse {
	if engine == "" {
		return datastoreTypes
	}

	var filteredDatastoreTypes []dbaas_v2.DatastoreTypeResponse
	for _, dt := range datastoreTypes {
		if dt.Engine == engine {
			filteredDatastoreTypes = append(filteredDatastoreTypes, dt)
		}
	}

	return filteredDatastoreTypes
}

func flattenDBaaSV2DatastoreTypes(datastoreTypes []dbaas_v2.DatastoreTypeResponse) []any {
	datastoreTypesList := make([]any, len(datastoreTypes))
	for i, datastoreType := range datastoreTypes {
		datastoreTypesMap := make(map[string]any)
		datastoreTypesMap["id"] = datastoreType.ID
		datastoreTypesMap["engine"] = datastoreType.Engine
		datastoreTypesMap["version"] = datastoreType.Version

		datastoreTypesList[i] = datastoreTypesMap
	}

	return datastoreTypesList
}

func expandDBaaSV2DatastoreTypeSearchFilter(filterSet *schema.Set) (dbaasV2DatastoreTypeSearchFilter, error) {
	filter := dbaasV2DatastoreTypeSearchFilter{}
	if filterSet.Len() == 0 {
		return filter, nil
	}

	resourceFilterMap := filterSet.List()[0].(map[string]any)

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
