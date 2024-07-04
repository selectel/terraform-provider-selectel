---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_flavor_v1"
sidebar_current: "docs-selectel-datasource-dbaas-flavor-v1"
description: |-
  Provides a list of flavors available in Selectel Managed Databases.
---

# selectel\_dbaas\_flavors_v1

Provides a list of flavors available in Managed Databases. For more information about available configurations, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/en/cloud/managed-databases/postgresql/configurations/), [PostgreSQL for 1C](https://docs.selectel.ru/en/cloud/managed-databases/postgresql-for-1c/configurations-1c/), [PostgreSQL TimescaleDB](https://docs.selectel.ru/en/cloud/managed-databases/timescaledb/configurations/), [MySQL semi-sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-semi-sync/configurations/), [MySQL sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-sync/configurations/), [Redis](https://docs.selectel.ru/en/cloud/managed-databases/redis/configurations/), and [Kafka](https://docs.selectel.ru/en/cloud/managed-databases/kafka/configurations/).

## Example Usage

```hcl
data "selectel_dbaas_flavor_v1" "flavor" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `filter` - (Optional) Values to filter available flavors:

  * `vcpus` - (Optional) Number of vCPUs.

  * `ram` - (Optional) Amount of RAM in MB.

  * `disk` - (Optional) Volume size in GB.

  * `fl_size` - (Optional) Line of flavors. Available values are `standard` (for the Standard, CPU, and Memory lines) and `high_freq` (for the HighFreq line). Learn more about available lines for [PostgreSQL](https://docs.selectel.ru/en/cloud/managed-databases/postgresql/configurations/), [PostgreSQL for 1C](https://docs.selectel.ru/en/cloud/managed-databases/postgresql-for-1c/configurations-1c/), [PostgreSQL TimescaleDB](https://docs.selectel.ru/en/cloud/managed-databases/timescaledb/configurations/), [MySQL semi-sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-semi-sync/configurations/), [MySQL sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-sync/configurations/), [Redis](https://docs.selectel.ru/en/cloud/managed-databases/redis/configurations/), and [Kafka](https://docs.selectel.ru/en/cloud/managed-databases/kafka/configurations/).

  * `datastore_type_id` - (Optional) Unique identifier of the datastore type.

## Attributes Reference

* `flavors` - List of available flavors.

  * `id` - Unique identifier of the flavor.

  * `name` - Flavor name.

  * `description` - Flavor description.

  * `vcpus` - Number of vCPUs.

  * `ram` - Amount of RAM in MB.

  * `disk` - Volume size in GB.

  * `fl_size` - Line of flavors.

  * `datastore_type_ids` - List of datastore types that support this flavor.