---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_configuration_parameter_v1"
sidebar_current: "docs-selectel-datasource-dbaas-configuration-parameter-v1"
description: |-
  Provides a list of configuration parameters available for Selectel Managed Databases.
---

# selectel\_dbaas\_configuration_parameter_v1

Provides a list of configuration parameters available for Managed Databases. For more information about configuration parameters, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/cloud/managed-databases/postgresql/settings/), [PostgreSQL for 1C](https://docs.selectel.ru/cloud/managed-databases/postgresql-for-1c/settings-1c/), [PostgreSQL TimescaleDB](https://docs.selectel.ru/cloud/managed-databases/timescaledb/settings/), [MySQL semi-sync](https://docs.selectel.ru/cloud/managed-databases/mysql-semi-sync/settings/), [MySQL sync](https://docs.selectel.ru/cloud/managed-databases/mysql-sync/settings/), [Redis](https://docs.selectel.ru/cloud/managed-databases/redis/eviction-policy/), and [Kafka](https://docs.selectel.ru/cloud/managed-databases/kafka/settings/).

## Example Usage

```hcl
data "selectel_dbaas_configuration_parameter_v1" "configuration_parameter_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#облачные-базы-данных).

* `filter` - (Optional) Values to filter available extensions.
  
  * `datastore_type_id` - (Optional) Unique identifier of the datastore type for which you get configuration parameters.  You can retrieve information about available datastore types with the [selectel_dbaas_datastore_type_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_datastore_type_v1) data source.

  * `name` - (Optional) Name of the configuration parameter to search.

## Attributes Reference

* `configuration_parameters` - List of  available configuration parameters.

  * `id` - Unique identifier of the configuration parameter.

  * `datastore_type_id` - Unique identifier of the datastore type for which the configuration parameter is available.

  * `name` - Name of the configuration parameter.

  * `type` - Type of the configuration parameter.

  * `unit` - Unit of the configuration parameter. Might be empty.

  * `min` - Minimum value of the configuration parameter. Might be empty.

  * `max` - Maximum value of the configuration parameter. Might be empty.

  * `default_value` - Default value of the configuration parameter. Might be empty.

  * `choices` - Available choices for the configuration parameter. Some parameters have list of available options.

  * `is_restart_required` - Shows if the database needs a restart to apply changes.

  * `is_changeable` - Shows if the parameter can be changed.