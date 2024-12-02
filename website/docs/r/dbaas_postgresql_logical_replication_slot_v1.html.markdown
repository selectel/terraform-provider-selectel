---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_postgresql_logical_replication_slot_v1"
sidebar_current: "docs-selectel-resource-dbaas-postgresql-logical-replication-slot-v1"
description: |-
  Creates and manages a logical replication slot in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_postgresql\_logical\_replication\_slot\_v1

Creates and manages a logical replication slot for Managed Databases using public API v1. Applicable to PostgreSQL and PostgreSQL TimescaleDB datastores. For more information about replication slots in Managed Databases, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/en/cloud/managed-databases/postgresql/replication-slots/) and [PostgreSQL TimescaleDB](https://docs.selectel.ru/en/cloud/managed-databases/timescaledb/replication-slots/).

## Example usage

```hcl
resource "selectel_dbaas_postgresql_logical_replication_slot_v1" "slot_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_postgresql_datastore_v1.datastore_1.id
  database_id  = selectel_dbaas_postgresql_database_v1.database_1.id
  name         = "test_slot"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new replication slot. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new replication slot. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `datastore_id` - (Required) Unique identifier of the associated datastore. Changing this creates a new replication slot. Retrieved from the [selectel_dbaas_postgresql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_datastore_v1)

* `database_id` - (Required) Unique identifier of the associated database. Changing this creates a new replication slot. Not applicable to a Redis datastore. Retrieved from the [selectel_dbaas_postgresql_database_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_database_v1) resource.

* `name` - (Required) Slot name. Can contain only lowercase letters, numbers, and an underscore. Changing this creates a new replication slot.

## Attributes Reference

* `status` - Status of the replication slot.

## Import

You can import a replication slot:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
export INFRA_REGION=<selectel_pool>
terraform import selectel_dbaas_postgresql_logical_replication_slot_v1.slot_1 <replication_slot_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.
  
* `<replication_slot_id>` — Unique identifier of the replication slot, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the replication slot ID, use [Selectel Cloud Management API](https://developers.selectel.ru/docs/selectel-cloud-platform/dbaas_api/).