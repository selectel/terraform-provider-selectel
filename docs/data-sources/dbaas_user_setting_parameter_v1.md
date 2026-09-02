---
page_title: "Selectel: selectel_dbaas_user_setting_parameter_v1"
description: |-
  Provides a list of user setting parameters available for Selectel Managed Databases.
---

# selectel\_dbaas\_user\_setting\_parameter\_v1

Provides a list of per-user setting parameters available for Managed Databases. Supported only for PostgreSQL, PostgreSQL for 1C, and PostgreSQL TimescaleDB. For more information about user settings, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/en/cloud/managed-databases/postgresql/manage-users/).

## Example Usage

```terraform
data "selectel_dbaas_user_setting_parameter_v1" "user_setting_parameter_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"

  filter {
    name = "statement_timeout"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `filter` - (Optional) Values to filter available user setting parameters.

  * `id` - (Optional) Unique identifier of the user setting parameter.

  * `name` - (Optional) Name of the user setting parameter to search.

  * `datastore_type_id` - (Optional) Unique identifier of the cluster type to list user setting parameters for. You can retrieve information about available cluster types with the [selectel_dbaas_datastore_type_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_datastore_type_v1) data source.

  * `datastore_id` - (Optional) Unique identifier of the cluster. Used to list user setting parameters for the cluster type of this cluster. Takes precedence over `datastore_type_id`. Retrieved from the [selectel_dbaas_postgresql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_datastore_v1) resource.

## Attributes Reference

* `user_setting_parameters` - List of available user setting parameters.

  * `id` - Unique identifier of the user setting parameter.

  * `name` - Name of the user setting parameter.

  * `type` - Type of the user setting parameter.

  * `unit` - Unit of the user setting parameter. Might be empty.

  * `min` - Minimum value of the user setting parameter. Might be empty.

  * `max` - Maximum value of the user setting parameter. Might be empty.

  * `default_value` - Default value of the user setting parameter. Might be empty.

  * `choices` - Available choices for the user setting parameter. Some parameters have a list of available options.

  * `apply_mechanism` - How the parameter is applied on the cluster: `guc` or `role_attr`.

  * `is_available_for_customer` - Shows if the parameter is available for the customer.

  * `is_changeable` - Shows if the parameter can be changed.

  * `can_be_empty` - Shows if the parameter value can be empty.