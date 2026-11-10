package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
)

type userSettingParameterSearchFilter struct {
	id              string
	name            string
	datastoreTypeID string
	datastoreID     string
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
						"default_value": {
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
						"can_be_empty": {
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

	filter := expandDBaaSUserSettingParameterSearchFilter(d.Get("filter").(*schema.Set))

	params := &dbaas.UserSettingParameterQueryParams{
		ID:              filter.id,
		Name:            filter.name,
		DatastoreTypeID: filter.datastoreTypeID,
		DatastoreID:     filter.datastoreID,
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

func expandDBaaSUserSettingParameterSearchFilter(filterSet *schema.Set) userSettingParameterSearchFilter {
	filter := userSettingParameterSearchFilter{}
	if filterSet.Len() == 0 {
		return filter
	}

	resourceFilterMap := filterSet.List()[0].(map[string]any)

	if id, ok := resourceFilterMap["id"]; ok {
		filter.id = id.(string)
	}
	if name, ok := resourceFilterMap["name"]; ok {
		filter.name = name.(string)
	}
	if datastoreTypeID, ok := resourceFilterMap["datastore_type_id"]; ok {
		filter.datastoreTypeID = datastoreTypeID.(string)
	}
	if datastoreID, ok := resourceFilterMap["datastore_id"]; ok {
		filter.datastoreID = datastoreID.(string)
	}

	return filter
}

func flattenDBaaSUserSettingParameters(userSettingParameters []dbaas.UserSettingParameter) []any {
	userSettingParametersList := make([]any, len(userSettingParameters))
	for i, param := range userSettingParameters {
		userSettingParametersMap := make(map[string]any)
		userSettingParametersMap["id"] = param.ID
		userSettingParametersMap["name"] = param.Name
		userSettingParametersMap["type"] = param.Type
		userSettingParametersMap["unit"] = param.Unit
		userSettingParametersMap["min"] = convertFieldToStringByType(param.Min)
		userSettingParametersMap["max"] = convertFieldToStringByType(param.Max)
		userSettingParametersMap["default_value"] = param.DefaultValue
		userSettingParametersMap["choices"] = convertListParametersTypes(param.Choices)
		userSettingParametersMap["apply_mechanism"] = param.ApplyMechanism
		userSettingParametersMap["is_available_for_customer"] = param.IsAvailableForCustomer
		userSettingParametersMap["is_changeable"] = param.IsChangeable
		userSettingParametersMap["can_be_empty"] = param.CanBeEmpty

		userSettingParametersList[i] = userSettingParametersMap
	}

	return userSettingParametersList
}
