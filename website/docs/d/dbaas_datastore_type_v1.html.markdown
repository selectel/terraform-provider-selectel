---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_datastore_type_v1"
sidebar_current: "docs-selectel-datasource-dbaas-datastore-type-v1"
description: |-
  Provides a list of available datastore types in Selectel Managed Databases.
---

# selectel\_dbaas\_datastore_type_v1

Provides a list of available datastore types in Managed Databases. For more information about available datastore types, see the [official Selectel documentation](https://docs.selectel.ru/cloud/managed-databases/about/about-managed-databases/#поддерживаемые-облачные-базы-данных).

## Example Usage for PostgreSQL

```hcl
data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine  = "postgresql"
    version = "14"
  }
}
```

## Example Usage for PostgreSQL for 1C

```hcl
data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine  = "postgresql"
    version = "14-1C"
  }
}
```

## Example Usage for PostgreSQL TimescaleDB

```hcl
data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine  = "postgresql"
    version = "14-TimescaleDB"
  }
}
```

## Example Usage for MySQL semi-sync

```hcl
data "selectel_dbaas_datastore_type_v1" "datastore_type_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine  = "mysql_native"
    version = "8"
  }
}
```

## Example Usage for MySQL sync

```hcl
data "selectel_dbaas_datastore_type_v1" "datastore_type_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine  = "mysql"
    version = "8"
  }
}
```

## Example Usage for Redis

```hcl
data "selectel_dbaas_datastore_type_v1" "datastore_type_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine  = "redis"
    version = "6"
  }
}
```

## Example Usage for Kafka

```hcl
data "selectel_dbaas_datastore_type_v1" "datastore_type_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine  = "kafka"
    version = "3.5"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#облачные-базы-данных).

* `filter` - (Optional) Values to filter available datastore types:

  * `engine` - (Optional) Engine of the datastore type to search. Available values are `postgresql` (for [PostgreSQL](https://docs.selectel.ru/cloud/managed-databases/postgresql/), [PostgreSQL for 1C](https://docs.selectel.ru/cloud/managed-databases/postgresql-for-1c/), [PostgreSQL TimescaleDB](https://docs.selectel.ru/cloud/managed-databases/timescaledb/)), `mysql` (for [MySQL sync](https://docs.selectel.ru/cloud/managed-databases/mysql-sync/)), `mysql_native` (for [MySQL semi-sync](https://docs.selectel.ru/cloud/managed-databases/mysql-semi-sync/)), and `redis`(for [Redis](https://docs.selectel.ru/cloud/managed-databases/redis/)).

  * `version` - (Optional) Version of the datastore type to search. For PostgreSQL for 1C, the versions are in the format `<version_number>-1C`. For PostgreSQL TimescaleDB, the versions are in the format `<version_number>-TimescaleDB`. Learn more about available versions for [PostgreSQL](https://docs.selectel.ru/cloud/managed-databases/postgresql/configurations/), [PostgreSQL for 1C](https://docs.selectel.ru/cloud/managed-databases/postgresql-for-1c/configurations-1c/), [PostgreSQL TimescaleDB](https://docs.selectel.ru/cloud/managed-databases/timescaledb/configurations/), [MySQL semi-sync](https://docs.selectel.ru/cloud/managed-databases/mysql-semi-sync/configurations/), [MySQL sync](https://docs.selectel.ru/cloud/managed-databases/mysql-sync/configurations/), and [Redis](https://docs.selectel.ru/cloud/managed-databases/redis/configurations/).

## Attributes Reference

* `datastore_types` - List of available datastore types.

  * `id` - ID of the datastore type.

  * `engine` - Engine of the datastore type.

  * `version` - Version of the datastore type.