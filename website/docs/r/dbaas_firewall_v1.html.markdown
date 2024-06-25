---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_firewall_v1"
sidebar_current: "docs-selectel-resource-dbaas-firewall-v1"
description: |-
  Creates and manages a list of IP-addresses with access to the datastore in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_firewall\_v1

Creates and manages a list of IP-addresses with access to a datastore in Managed Databases using public API v1. For more information about a firewall, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/en/cloud/managed-databases/postgresql/network-access-control/), [PostgreSQL for 1C](https://docs.selectel.ru/en/cloud/managed-databases/postgresql-for-1c/network-access-control-1c/), [PostgreSQL TimescaleDB](https://docs.selectel.ru/en/cloud/managed-databases/timescaledb/network-access-control/), [MySQL sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-sync/network-access-control/), [MySQL semi-sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-semi-sync/network-access-control/), [Kafka](https://docs.selectel.ru/en/cloud/managed-databases/kafka/network-access-control/), and [Redis](https://docs.selectel.ru/en/cloud/managed-databases/redis/network-access-control/).

## Example usage for PostgreSQL, PostgreSQL TimescaleDB, PostgreSQL for 1C

```hcl
resource "selectel_dbaas_firewall_v1" "firewall_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_postgresql_datastore_v1.datastore_1.id
  ips          = [ "127.0.0.1" ]
}
```

## Example usage for MySQL semi-sync and MySQL sync

```hcl
resource "selectel_dbaas_firewall_v1" "firewall_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_mysql_datastore_v1.datastore_1.id
  ips          = [ "127.0.0.1" ]
}
```

## Example usage for Redis

```hcl
resource "selectel_dbaas_firewall_v1" "firewall_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_redis_datastore_v1.datastore_1.id
  ips          = [ "127.0.0.1" ]
}
```

## Example usage for Kafka

```hcl
resource "selectel_dbaas_firewall_v1" "firewall_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_redis_datastore_v1.datastore_1.id
  ips          = [ "127.0.0.1" ]
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new datastore. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new datastore. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `datastore_id` - (Required) Unique identifier of the associated datastore. Changing this updates the list of IP-addresses with access to the datastore. Retrieved from the [selectel_dbaas_postgresql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_datastore_v1), [selectel_dbaas_mysql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_mysql_datastore_v1), [selectel_dbaas_redis_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_redis_datastore_v1) or [selectel_dbaas_kafka_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_kafka_datastore_v1) resource depending on the datastore type you use.

* `ips` - (Required) List of IP-addresses with access to the datastore.
