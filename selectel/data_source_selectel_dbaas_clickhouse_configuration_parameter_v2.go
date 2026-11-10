package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dbaas_v2_ch "github.com/selectel/dbaas-go/v2/clickhouse"
)

func dataSourceDBaaSV2ClickhouseConfigurationParameter() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDBaaSV2ClickhouseConfigurationParameterRead,
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
						"can_be_empty": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"is_multiple_choice_available": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func filterDBaaSV2ClickhouseConfigurationParametersByDatastoreTypeID(
	configurationParameters []dbaas_v2_ch.DatastoreConfigurationParameterResponse, datastoreTypeID string,
) []dbaas_v2_ch.DatastoreConfigurationParameterResponse {
	if datastoreTypeID == "" {
		return configurationParameters
	}

	var filteredConfigurationParameters []dbaas_v2_ch.DatastoreConfigurationParameterResponse
	for _, param := range configurationParameters {
		if param.DatastoreTypeID == datastoreTypeID {
			filteredConfigurationParameters = append(filteredConfigurationParameters, param)
		}
	}

	return filteredConfigurationParameters
}

func filterDBaaSV2ClickhouseConfigurationParametersByName(
	configurationParameters []dbaas_v2_ch.DatastoreConfigurationParameterResponse, name string,
) []dbaas_v2_ch.DatastoreConfigurationParameterResponse {
	if name == "" {
		return configurationParameters
	}

	var filteredConfigurationParameters []dbaas_v2_ch.DatastoreConfigurationParameterResponse
	for _, param := range configurationParameters {
		if param.Name == name {
			filteredConfigurationParameters = append(filteredConfigurationParameters, param)
			break
		}
	}

	return filteredConfigurationParameters
}

func flattenDBaaSV2ClickhouseConfigurationParameters(configurationParameters []dbaas_v2_ch.DatastoreConfigurationParameterResponse) []any {
	configurationParametersList := make([]any, len(configurationParameters))
	for i, param := range configurationParameters {
		configurationParametersMap := make(map[string]any)
		configurationParametersMap["id"] = param.ID
		configurationParametersMap["datastore_type_id"] = param.DatastoreTypeID
		configurationParametersMap["name"] = param.Name
		configurationParametersMap["type"] = param.Type
		configurationParametersMap["min"] = param.MinValue
		configurationParametersMap["max"] = param.MaxValue
		configurationParametersMap["default_value"] = convertFieldToStringByType(param.DefaultValue)
		configurationParametersMap["choices"] = param.Choices
		configurationParametersMap["invalid_values"] = param.InvalidValues
		configurationParametersMap["is_restart_required"] = param.IsRestartRequired
		configurationParametersMap["is_changeable"] = param.IsChangeable
		configurationParametersMap["is_multiple_choice_available"] = param.IsMultipleChoiceAvailable
		configurationParametersMap["can_be_empty"] = param.CanBeEmpty

		configurationParametersList[i] = configurationParametersMap
	}

	return configurationParametersList
}

func dataSourceDBaaSV2ClickhouseConfigurationParameterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	configurationParameters, err := dbaasClient.ClickHouse.GetDatastoreConfigurationParameters(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectConfigurationParameters, err))
	}

	configurationParametersIDs := make([]string, 0, len(configurationParameters))
	for _, param := range configurationParameters {
		configurationParametersIDs = append(configurationParametersIDs, param.ID)
	}

	filter := expandDBaaSV2ConfigurationParameterSearchFilter(d.Get("filter").(*schema.Set))

	configurationParameters = filterDBaaSV2ClickhouseConfigurationParametersByDatastoreTypeID(configurationParameters, filter.datastoreTypeID)
	configurationParameters = filterDBaaSV2ClickhouseConfigurationParametersByName(configurationParameters, filter.name)

	configurationParametersFlatten := flattenDBaaSV2ClickhouseConfigurationParameters(configurationParameters)
	if err := d.Set("configuration_parameters", configurationParametersFlatten); err != nil {
		return diag.FromErr(err)
	}
	checksum, err := stringListChecksum(configurationParametersIDs)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}
