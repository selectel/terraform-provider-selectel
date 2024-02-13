---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_grant_v1"
sidebar_current: "docs-selectel-resource-dbaas-grant-v1"
description: |-
  Grants privileges to the users in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_grant\_v1

Grants privileges to the users in Managed Databases using public API v1. Not applicable to Redis and Kafka. Learn more about Managed Databases in the [official Selectel documentation](https://docs.selectel.ru/cloud/managed-databases/).

## Example usage

### PostgreSQL, PostgreSQL for 1C, and PostgreSQL TimescaleDB

```hcl
resource "selectel_dbaas_grant_v1" "grant_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_postgresql_datastore_v1.datastore_1.id
  database_id  = selectel_dbaas_postgresql_database_v1.database_1.id
  user_id      = selectel_dbaas_user_v1.user_1.id
}
```

### MySQL semi-sync and MySQL sync

```hcl
resource "selectel_dbaas_grant_v1" "grant_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_mysql_datastore_v1.datastore_1.id
  database_id  = selectel_dbaas_mysql_database_v1.database_1.id
  user_id      = selectel_dbaas_user_v1.user_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new privilege for the user. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new privilege for the user.

* `datastore_id` - (Required) Unique identifier of the associated datastore. Changing this creates a new privilege for the user. Retrieved from the [selectel_dbaas_postgresql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_datastore_v1) or [selectel_dbaas_mysql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_mysql_datastore_v1) resource depending on the datastore type you use.

* `database_id` - (Required) Unique identifier of the associated database. Changing this creates a new privilege for the user. Retrieved from the [selectel_dbaas_postgresql_database_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_database_v1) or [selectel_dbaas_mysql_database_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_mysql_database_v1) resource depending on the datastore type you use.

* `user_id` - (Required) Unique identifier of the associated user. Changing this creates a new privilege for the user. Retrieved from the [selectel_dbaas_user_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_user_v1) resource.

## Attributes Reference

* `status` - Status of the user privilege.