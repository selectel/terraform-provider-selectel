package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
)

type configurationParameterSearchFilter struct {
	datastoreTypeID string
	name            string
}

func dataSourceDBaaSConfigurationParameterV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDBaaSConfigurationParameterV1Read,
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
						"datastore_type_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"configuration_parameters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"datastore_type_id": {
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
						"invalid_values": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"is_restart_required": {
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

func dataSourceDBaaSConfigurationParameterV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	configurationParameters, err := dbaasClient.ConfigurationParameters(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectConfigurationParameters, err))
	}

	configurationParametersIDs := make([]string, 0, len(configurationParameters))
	for _, param := range configurationParameters {
		configurationParametersIDs = append(configurationParametersIDs, param.ID)
	}

	filter, err := expandConfigurationParameterSearchFilter(d.Get("filter").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	configurationParameters = filterConfigurationParametersByDatastoreTypeID(configurationParameters, filter.datastoreTypeID)
	configurationParameters = filterConfigurationParametersByName(configurationParameters, filter.name)

	configurationParametersFlatter := flattenDBaaSConfigurationParameters(configurationParameters)
	if err := d.Set("configuration_parameters", configurationParametersFlatter); err != nil {
		return diag.FromErr(err)
	}
	checksum, err := stringListChecksum(configurationParametersIDs)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}

func expandConfigurationParameterSearchFilter(filterSet *schema.Set) (configurationParameterSearchFilter, error) {
	filter := configurationParameterSearchFilter{}
	if filterSet.Len() == 0 {
		return filter, nil
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	datastoreTypeID, ok := resourceFilterMap["datastore_type_id"]
	if ok {
		filter.datastoreTypeID = datastoreTypeID.(string)
	}

	name, ok := resourceFilterMap["name"]
	if ok {
		filter.name = name.(string)
	}

	return filter, nil
}

func filterConfigurationParametersByDatastoreTypeID(configurationParameters []dbaas.ConfigurationParameter, datastoreTypeID string) []dbaas.ConfigurationParameter {
	if datastoreTypeID == "" {
		return configurationParameters
	}

	var filteredConfigurationParameters []dbaas.ConfigurationParameter
	for _, param := range configurationParameters {
		if param.DatastoreTypeID == datastoreTypeID {
			filteredConfigurationParameters = append(filteredConfigurationParameters, param)
		}
	}

	return filteredConfigurationParameters
}

func filterConfigurationParametersByName(configurationParameters []dbaas.ConfigurationParameter, name string) []dbaas.ConfigurationParameter {
	if name == "" {
		return configurationParameters
	}

	var filteredConfigurationParameters []dbaas.ConfigurationParameter
	for _, param := range configurationParameters {
		if param.Name == name {
			filteredConfigurationParameters = append(filteredConfigurationParameters, param)
		}
	}

	return filteredConfigurationParameters
}

func convertListParametersTypes(parameters []interface{}) []string {
	parameterList := make([]string, len(parameters))
	for i, value := range parameters {
		parameterList[i] = convertFieldToStringByType(value)
	}

	return parameterList
}

func flattenDBaaSConfigurationParameters(configurationParameters []dbaas.ConfigurationParameter) []interface{} {
	configurationParametersList := make([]interface{}, len(configurationParameters))
	for i, param := range configurationParameters {
		configurationParametersMap := make(map[string]interface{})
		configurationParametersMap["id"] = param.ID
		configurationParametersMap["datastore_type_id"] = param.DatastoreTypeID
		configurationParametersMap["name"] = param.Name
		configurationParametersMap["type"] = param.Type
		configurationParametersMap["unit"] = param.Unit
		configurationParametersMap["min"] = convertFieldToStringByType(param.Min)
		configurationParametersMap["max"] = convertFieldToStringByType(param.Max)
		configurationParametersMap["default_value"] = convertFieldToStringByType(param.DefaultValue)
		configurationParametersMap["choices"] = convertListParametersTypes(param.Choices)
		configurationParametersMap["invalid_values"] = convertListParametersTypes(param.InvalidValues)
		configurationParametersMap["is_restart_required"] = param.IsRestartRequired
		configurationParametersMap["is_changeable"] = param.IsChangeable

		configurationParametersList[i] = configurationParametersMap
	}

	return configurationParametersList
}
