package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
)

type userSettingParameterSearchFilter struct {
	id                 string
	name               string
	datastoreGroupName string
	datastoreTypeID    string
	datastoreID        string
}

func dataSourceDBaaSUserSettingParameterV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDBaaSUserSettingParameterV1Read,
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
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"datastore_group_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"datastore_type_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"datastore_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"user_setting_parameters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"datastore_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"unit": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"min": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"max": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"choices": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"apply_mechanism": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_available_for_customer": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"is_changeable": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDBaaSUserSettingParameterV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter, err := expandUserSettingParameterSearchFilter(d.Get("filter").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	params := &dbaas.UserSettingParameterQueryParams{
		ID:                 filter.id,
		Name:               filter.name,
		DatastoreGroupName: filter.datastoreGroupName,
		DatastoreTypeID:    filter.datastoreTypeID,
		DatastoreID:        filter.datastoreID,
	}

	userSettingParameters, err := dbaasClient.UserSettingParameters(ctx, params)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectUserSettingParameters, err))
	}

	userSettingParametersIDs := make([]string, 0, len(userSettingParameters))
	for _, param := range userSettingParameters {
		userSettingParametersIDs = append(userSettingParametersIDs, param.ID)
	}

	if err := d.Set("user_setting_parameters", flattenDBaaSUserSettingParameters(userSettingParameters)); err != nil {
		return diag.FromErr(err)
	}
	checksum, err := stringListChecksum(userSettingParametersIDs)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}

func expandUserSettingParameterSearchFilter(filterSet *schema.Set) (userSettingParameterSearchFilter, error) {
	filter := userSettingParameterSearchFilter{}
	if filterSet.Len() == 0 {
		return filter, nil
	}

	resourceFilterMap := filterSet.List()[0].(map[string]any)

	if id, ok := resourceFilterMap["id"]; ok {
		filter.id = id.(string)
	}
	if name, ok := resourceFilterMap["name"]; ok {
		filter.name = name.(string)
	}
	if groupName, ok := resourceFilterMap["datastore_group_name"]; ok {
		filter.datastoreGroupName = groupName.(string)
	}
	if datastoreTypeID, ok := resourceFilterMap["datastore_type_id"]; ok {
		filter.datastoreTypeID = datastoreTypeID.(string)
	}
	if datastoreID, ok := resourceFilterMap["datastore_id"]; ok {
		filter.datastoreID = datastoreID.(string)
	}

	return filter, nil
}

func flattenDBaaSUserSettingParameters(userSettingParameters []dbaas.UserSettingParameter) []any {
	userSettingParametersList := make([]any, len(userSettingParameters))
	for i, param := range userSettingParameters {
		userSettingParametersMap := make(map[string]any)
		userSettingParametersMap["id"] = param.ID
		userSettingParametersMap["datastore_group_name"] = param.DatastoreGroupName
		userSettingParametersMap["name"] = param.Name
		userSettingParametersMap["type"] = param.Type
		userSettingParametersMap["unit"] = param.Unit
		userSettingParametersMap["min"] = convertFieldToStringByType(param.Min)
		userSettingParametersMap["max"] = convertFieldToStringByType(param.Max)
		userSettingParametersMap["choices"] = convertListParametersTypes(param.Choices)
		userSettingParametersMap["apply_mechanism"] = param.ApplyMechanism
		userSettingParametersMap["is_available_for_customer"] = param.IsAvailableForCustomer
		userSettingParametersMap["is_changeable"] = param.IsChangeable

		userSettingParametersList[i] = userSettingParametersMap
	}

	return userSettingParametersList
}
