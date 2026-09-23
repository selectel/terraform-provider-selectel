---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_datastore_type_v2"
sidebar_current: "docs-selectel-datasource-dbaas-datastore-type-v2"
description: |-
  Provides a list of available cluster types in Selectel Managed Databases using public API v2.
---

# selectel\_dbaas\_datastore\_type\_v2

Provides a list of available cluster types in Managed Databases using public API v2. For more information about available cluster types, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-databases/about/about-managed-databases/#supported-databases).

## Example Usage

### ClickHouse

```hcl
data "selectel_dbaas_datastore_type_v2" "dt" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine = "clickhouse"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `filter` - (Optional) Values to filter available cluster types.

  * `engine` - (Optional) Engine of the cluster type to search. Available values: `clickhouse`, `kafka`, `mysql`, `mysql_native`, `postgresql`, `redis`.

  * `version` - (Optional) Version of the cluster type to search.

## Attributes Reference

* `datastore_types` - List of available cluster types.

  * `id` - Unique identifier of the cluster type.

  * `engine` - Engine of the cluster type.

  * `version` - Version of the cluster type.
