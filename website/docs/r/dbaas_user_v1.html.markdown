---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_user_v1"
sidebar_current: "docs-selectel-resource-dbaas-user-v1"
description: |-
  Creates and manages a user in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_user\_v1

Creates and manages a user in Managed Databases using public API v1. Not applicable to Redis. For more information about managing users in Managed Databases, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/cloud/managed-databases/postgresql/manage-users/), [PostgreSQL for 1C](https://docs.selectel.ru/cloud/managed-databases/postgresql-for-1c/manage-users-1c/), [PostgreSQL TimescaleDB](https://docs.selectel.ru/cloud/managed-databases/timescaledb/manage-users/), [MySQL semi-sync](https://docs.selectel.ru/cloud/managed-databases/mysql-semi-sync/manage-users/), [MySQL sync](https://docs.selectel.ru/cloud/managed-databases/mysql-sync/manage-users/), and [Kafka](https://docs.selectel.ru/cloud/managed-databases/kafka/manage-users/).

## Example usage

### PostgreSQL, PostgreSQL for 1C, and PostgreSQL TimescaleDB

```hcl
resource "selectel_dbaas_user_v1" "user_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_postgresql_datastore_v1.datastore_1.id
  name         = "user"
  password     = "secret"
}
```

### MySQL semi-sync and MySQL sync

```hcl
resource "selectel_dbaas_user_v1" "user_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_mysql_datastore_v1.datastore_1.id
  name         = "user"
  password     = "secret"
}
```

### Kafka

```hcl
resource "selectel_dbaas_user_v1" "user_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_kafka_datastore_v1.datastore_1.id
  name         = "user"
  password     = "secret"
}
```

## Argument Reference

* `name` - (Required, Sensitive) User name. Changing this creates a new user.

* `password` - (Required, Sensitive) User password.

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new user. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new user. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#облачные-базы-данных).

* `datastore_id` - (Required) Unique identifier of the associated datastore. Changing this creates a new user. Retrieved from the [selectel_dbaas_postgresql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_datastore_v1) or [selectel_dbaas_mysql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_mysql_datastore_v1) resource depending on the datastore type you use.

## Attributes Reference

* `status` - User status.

## Import

You can import a user:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export SEL_PROJECT_ID=<selectel_project_id>
export SEL_REGION=<selectel_pool>
terraform import selectel_dbaas_user_v1.user_1 <user_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<user_id>` — Unique identifier of the user, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the user ID in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ the cluster page ⟶ the **Users** tab. The user ID is under the user name.