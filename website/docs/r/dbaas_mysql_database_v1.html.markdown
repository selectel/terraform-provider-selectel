---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_mysql_database_v1"
sidebar_current: "docs-selectel-resource-dbaas-mysql-database-v1"
description: |-
  Creates and manages a MySQL database in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_mysql\_database\_v1

Creates and manages a MySQL database using public API v1. Applicable to MySQL sync and MySQL semi-sync datastores, the type is determined by the [selectel_dbaas_mysql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_mysql_datastore_v1) resource. For more information about MySQL databases, see the official Selectel documentation for [MySQL sync](https://docs.selectel.ru/cloud/managed-databases/mysql-sync/) and [MySQL semi-sync](https://docs.selectel.ru/cloud/managed-databases/mysql-semi-sync/).

## Example usage

```hcl
resource "selectel_dbaas_mysql_database_v1" "database_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_mysql_datastore_v1.datastore_1.id
  name         = "database_1"
}
```

## Argument Reference

* `name` - (Required) Database name. Changing this creates a new database.

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new database. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new database. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#облачные-базы-данных).

* `datastore_id` - (Required) Unique identifier of the associated datastore. Changing this creates a new database. Retrieved from the [selectel_dbaas_mysql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_mysql_datastore_v1) resource.

## Attributes Reference

* `status` - Database status.

## Import

You can import a database:

```shell
<<<<<<< HEAD
terraform import selectel_dbaas_mysql_database_v1.database_1 <database_id>
=======
<<<<<<< HEAD
terraform import selectel_dbaas_mysql_database_v1.database_1 <database_id>
=======
$ export OS_DOMAIN_NAME=999999
$ export OS_USERNAME=example_user
$ export OS_PASSWORD=example_password
$ export SEL_PROJECT_ID=SELECTEL_VPC_PROJECT_ID
$ export SEL_REGION=SELECTEL_VPC_REGION
$ terraform import selectel_dbaas_database_v1.database_1 b311ce58-2658-46b5-b733-7a0f418703f2
>>>>>>> ceb748d (Move domains resources to keystone auth)
>>>>>>> upstream/master
```

where `<database_id>` is a unique identifier of the database, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the database ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ the cluster page ⟶ the **Databases** tab ⟶ copy the ID under the database name.

### Environment Variables

For import, you must set environment variables:

* `SEL_TOKEN=<selectel_api_token>`

* `SEL_PROJECT_ID=<selectel_project_id>`

* `SEL_REGION=<selectel_pool>`

where:

* `<selectel_api_token>` — Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶   **API keys**  ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to Cloud Platform ⟶ project name ⟶  copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-kubernetes/about/projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.