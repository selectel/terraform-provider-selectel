---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_flavor_v2"
sidebar_current: "docs-selectel-datasource-dbaas-flavor-v2"
description: |-
  Provides a list of flavors available in Selectel Managed Databases using public API v2.
---

# selectel\_dbaas\_flavor\_v2

Provides a list of flavors available in Managed Databases using public API v2. Learn more about available configurations for [ClickHouse](https://docs.selectel.ru/en/cloud/managed-databases/clickhouse/configurations/).

## Example Usage

```hcl
data "selectel_dbaas_flavor_v2" "flavor" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    datastore_type_id = data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id
    allowed_role      = "KEEPER"
    fl_size           = "STANDARD"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `filter` - (Optional) Values to filter available flavors.

  * `vcpus` - (Optional) Number of vCPUs.

  * `ram` - (Optional) Amount of RAM in MB.

  * `disk` - (Optional) Volume size in GB.

  * `fl_size` - (Optional) Line of flavors. Available values are `STANDARD` (for the Standard, CPU, and Memory lines) and `HIGH_FREQ` (for the HighFreq line). Learn more about available configurations for [ClickHouse](https://docs.selectel.ru/en/cloud/managed-databases/clickhouse/configurations/).

  * `datastore_type_id` - (Optional) Unique identifier of the cluster type. You can retrieve information about available cluster types with the [selectel_dbaas_datastore_type_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_datastore_type_v2) data source.

  * `allowed_role` - (Optional) Role that the flavor supports. Available values are `DATA` and `KEEPER`. You can use this filter, for example, to find flavors for ClickHouse Keeper nodes.

## Attributes Reference

* `flavors` - List of available flavors.

  * `id` - Unique identifier of the flavor.

  * `vcpus` - Number of vCPUs.

  * `ram` - Amount of RAM in MB.

  * `disk` - Volume size in GB.

  * `fl_size` - Line of flavors.

  * `datastore_type_ids` - List of cluster types that support this flavor.

  * `allowed_roles` - List of roles that the flavor supports.
