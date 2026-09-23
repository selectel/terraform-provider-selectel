---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_clickhouse_shard_group_v2"
sidebar_current: "docs-selectel-resource-dbaas-clickhouse-shard-group-v2"
description: |-
  Creates and manages a ClickHouse shard group in Selectel Managed Databases using public API v2.
---

# selectel\_dbaas\_clickhouse\_shard\_group\_v2

Creates and manages a ClickHouse shard group using public API v2. Shard groups allow you to combine shards (node groups with `DATA` role) into logical groups for distributed queries. For more information about Managed Databases, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-databases/clickhouse/).

## Example usage

```hcl
resource "selectel_dbaas_clickhouse_shard_group_v2" "shard_group_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_clickhouse_datastore_v2.cluster_1.id
  name         = "group-1"
  description  = "First shard group"
  shard_names  = ["shard1", "shard2"]
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new shard group. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the cluster is located, for example, `ru-3`. Changing this creates a new shard group. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `datastore_id` - (Required) Unique identifier of the associated ClickHouse cluster. Changing this creates a new shard group. Retrieved from the [selectel_dbaas_clickhouse_datastore_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_clickhouse_datastore_v2) resource.

* `name` - (Required) Name of the shard group. Changing this creates a new shard group.

* `shard_names` - (Required) List of shard names to include in the group. Each shard name must correspond to a node group with the `DATA` role in the cluster. At least one shard must be specified.

* `description` - (Optional) Description of the shard group.

## Attributes Reference

* `id` - Unique identifier of the shard group.

## Import

You can import a shard group:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
export INFRA_REGION=<selectel_pool>
terraform import selectel_dbaas_clickhouse_shard_group_v2.shard_group_1 <datastore_id>/<shard_group_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<datastore_id>` — Unique identifier of the cluster. To get the cluster ID in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ copy the ID under the cluster name.

* `<shard_group_id>` — Unique identifier of the shard group.
