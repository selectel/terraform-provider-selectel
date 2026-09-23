---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_clickhouse_configuration_parameter_v2"
sidebar_current: "docs-selectel-datasource-dbaas-clickhouse-configuration-parameter-v2"
description: |-
  Provides a list of configuration parameters available for Selectel Managed ClickHouse clusters.
---

# selectel\_dbaas\_clickhouse\_configuration\_parameter\_v2

Provides a list of configuration parameters available for Managed ClickHouse clusters. For more information about configuration parameters, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-databases/clickhouse/settings/).

## Example Usage

```hcl
data "selectel_dbaas_clickhouse_configuration_parameter_v2" "configuration_parameter_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `filter` - (Optional) Values to filter available configuration parameters.

  * `datastore_type_id` - (Optional) Unique identifier of the cluster type for which you get configuration parameters. You can retrieve information about available cluster types with the [selectel_dbaas_datastore_type_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_datastore_type_v2) data source.

  * `name` - (Optional) Name of the configuration parameter to search.

## Attributes Reference

* `configuration_parameters` - List of available configuration parameters.

  * `id` - Unique identifier of the configuration parameter.

  * `datastore_type_id` - Unique identifier of the cluster type for which the configuration parameter is available.

  * `name` - Name of the configuration parameter.

  * `type` - Type of the configuration parameter.

  * `min` - Minimum value of the configuration parameter. Might be empty.

  * `max` - Maximum value of the configuration parameter. Might be empty.

  * `default_value` - Default value of the configuration parameter. Might be empty.

  * `choices` - Available choices for the configuration parameter. Some parameters have a list of available options.

  * `invalid_values` - Invalid values for the configuration parameter. Some parameters have a list of values within a range that are not available for the parameter.

  * `is_restart_required` - Shows if the database needs a restart to apply changes.

  * `is_changeable` - Shows if the parameter can be changed.

  * `can_be_empty` - Shows if the parameter value can be empty.

  * `is_multiple_choice_available` - Shows if the parameter supports multiple choices.
