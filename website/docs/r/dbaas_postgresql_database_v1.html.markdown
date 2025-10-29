---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_postgresql_database_v1"
sidebar_current: "docs-selectel-resource-dbaas-postgresql-database-v1"
description: |-
  Creates and manages a PostgreSQL database in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_postgresql\_database\_v1

Creates and manages a PostgreSQL database using public API v1. Applicable to PostgreSQL, PostgreSQL for 1C, and PostgreSQL TimescaleDB databases. The type is determined by the [selectel_dbaas_postgresql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_datastore_v1) resource. For more information about Managed Databases, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/en/cloud/managed-databases/postgresql/), [PostgreSQL for 1C](https://docs.selectel.ru/en/cloud/managed-databases/postgresql-for-1c/), and [PostgreSQL TimescaleDB](https://docs.selectel.ru/en/cloud/managed-databases/timescaledb/).

## Example usage

```hcl
resource "selectel_dbaas_postgresql_database_v1" "database_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_postgresql_datastore_v1.cluster_1.id
  owner_id     = selectel_dbaas_user_v1.user_1.id
  name         = "database_1"
}
```

## Argument Reference

* `name` - (Required) Database name. Changing this creates a new database.

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new database. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new database. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `datastore_id` - (Required) Unique identifier of the associated cluster. Changing this creates a new database. Retrieved from the [selectel_dbaas_mysql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_mysql_datastore_v1) resource.

* `owner_id` - (Required) Unique identifier of the database owner. Retrieved from the [selectel_dbaas_user_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_user_v1) resource.

* `lc_collate` - (Optional) LC_COLLATE setting for the PostgreSQL cluster. Determined when you create a database and you cannot change it afterwards. The default value is `C`.
  
  To get a list of available values, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**⟶ the cluster page ⟶ the **Databases** tab ⟶ **Create** ⟶ the **LC_COLLATE** field. The list of appropriate values is in the drop-down list.

* `lc_ctype` - (Optional) LC_CTYPE setting for the PostgreSQL cluster. Determined when you create a database and you cannot change it afterwards. The default value is `C`.

  To get a list of available values, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ the cluster page ⟶ the **Databases** tab ⟶ **Create** ⟶ the **LC_CTYPE** field. The list of appropriate values is in the drop-down list.

## Attributes Reference

* `status` - Database status.

## Import

You can import a database:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
export INFRA_REGION=<selectel_pool>
terraform import selectel_dbaas_postgresql_database_v1.database_1 <database_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<database_id>` — Unique identifier of the database, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the database ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ the cluster page ⟶ the **Databases** tab ⟶ copy the ID under the database name.
