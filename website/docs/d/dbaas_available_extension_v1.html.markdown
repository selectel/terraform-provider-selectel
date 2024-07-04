---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_available_extension_v1"
sidebar_current: "docs-selectel-datasource-dbaas-available-extension-v1"
description: |-
  Provides a list of extensions available for Selectel Managed Databases.
---

# selectel\_dbaas\_available_extension_v1

Provides a list of extensions available for Managed Databases. Applicable to PostgreSQL and PostgreSQL TimescaleDB. For more information about extensions, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/en/cloud/managed-databases/postgresql/add-extensions/) and [PostgreSQL TimescaleDB](https://docs.selectel.ru/en/cloud/managed-databases/timescaledb/add-extensions/).

## Example Usage

```hcl
data "selectel_dbaas_available_extension_v1" "available_extension_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `filter` - (Optional) Values to filter available extensions.

  * `name` - (Optional) Name of the extension to search.

## Attributes Reference

* `available_extensions` - List of the available extensions:

  * `id` - Unique identifier of the extension.

  * `name` - Extension name.

  * `datastore_type_ids` - List of datastore types that support the extension.

  * `dependency_ids` - List of extensions that depend on this extension.