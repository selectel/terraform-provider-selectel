---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_postgresql_logical_replication_slot_v1"
sidebar_current: "docs-selectel-resource-dbaas-postgresql-logical-replication-slot-v1"
description: |-
  Creates and manages a logical replication slot in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_postgresql\_logical\_replication\_slot\_v1

Creates and manages a logical replication slot for Managed Databases using public API v1. Applicable to PostgreSQL and PostgreSQL TimescaleDB  datastores. For more information about replication slots in Managed Databases, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/cloud/managed-databases/postgresql/replication-slots/) and [PostgreSQL TimescaleDB](https://docs.selectel.ru/cloud/managed-databases/timescaledb/replication-slots/).

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

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new replication slot. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new replication slot. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#облачные-базы-данных).

* `datastore_id` - (Required) Unique identifier of the associated datastore. Changing this creates a new replication slot. Retrieved from the [selectel_dbaas_postgresql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_datastore_v1)

* `database_id` - (Required) Unique identifier of the associated database. Changing this creates a new replication slot. Not applicable to a Redis datastore. Retrieved from the [selectel_dbaas_postgresql_database_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_database_v1) resource.

* `name` - (Required) Slot name. Can contain only lowercase letters, numbers, and an underscore. Changing this creates a new replication slot.

## Attributes Reference

* `status` - Status of the replication slot.

## Import

You can import a replication slot:

```shell
terraform import selectel_dbaas_postgresql_logical_replication_slot_v1.slot_1 <replication_slot_id>
```

where `<replication_slot_id>` is a unique identifier of the replication slot, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the replication slot ID, use [Selectel Cloud Management API](https://developers.selectel.ru/docs/selectel-cloud-platform/dbaas_api/).

### Environment Variables

For import, you must set environment variables:

* `SEL_TOKEN=<selectel_api_token>`

* `SEL_PROJECT_ID=<selectel_project_id>`

* `SEL_REGION=<selectel_pool>`

where:

* `<selectel_api_token>` — Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶   **API keys**  ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to Cloud Platform ⟶ project name ⟶  copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-kubernetes/about/projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.