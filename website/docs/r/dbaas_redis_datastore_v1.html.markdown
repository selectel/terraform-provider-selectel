---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_redis_datastore_v1"
sidebar_current: "docs-selectel-resource-dbaas-redis-datastore-v1"
description: |-
  Creates and manages a Redis datastore in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_redis\_datastore\_v1

Creates and manages a Redis datastore using public API v1. For more information about Managed Databases, see the [official Selectel documentation](https://docs.selectel.ru/cloud/managed-databases/redis/).

## Example usage

```hcl
resource "selectel_dbaas_redis_datastore_v1" "datastore_1" {
  name           = "datastore-1"
  project_id     = selectel_vpc_project_v2.project_1.id
  region         = "ru-3"
  type_id        = data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id
  subnet_id      = selectel_vpc_subnet_v2.subnet.subnet_id
  node_count     = 3
  flavor_id      = data.selectel_dbaas_flavor_v1.flavor.flavors[0].id
  redis_password = "secret"
}
```

## Argument Reference

* `name` - (Required) Datastore name. Changing this creates a new datastore.

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new datastore. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new datastore. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#managed-databases).

* `subnet_id` - (Required) Unique identifier of the associated OpenStack network. Changing this creates a new datastore. Learn more about the [openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_network_v2) resource in the official OpenStack documentation.

* `type_id` - (Required) Unique identifier of the datastore type. Changing this creates a new datastore. Retrieved from the [selectel_dbaas_datastore_type_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_datastore_type_v1) data source.

* `node_count` - (Required) Number of replicas in the datastore. Available values are `1` and `2`. Learn more about [Replication](https://docs.selectel.ru/cloud/managed-databases/about/about-managed-databases/#fault-tolerance-and-replication).

* `flavor_id` - (Required) Unique identifier of the flavor for the datastore. Retrieved from the [selectel_dbaas_flavor_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_flavor_v1) data source.

* `firewall` - (Optional) List of IP-addresses with access to the datastore

* `restore` - (Optional) Restores parameters for the datastore. Changing this creates a new datastore.

  * `datastore_id` - (Optional) Unique identifier of the datastore from which you restore. To get the datastore ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ copy the ID under the cluster name.
  
  * `target_time` - (Optional) Time within seven previous days when you have the datastore state to restore.

* `config` - (Optional) Configuration parameters for the datastore. You can retrieve information about available configuration parameters with the [selectel_dbaas_configuration_parameter_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_configuration_parameter_v1) data source.

* `redis_password` - (Required, Sensitive) Datastore password.

## Attributes Reference

* `status` - Datastore status.

* `connections` - DNS addresses to connect to the datastore.

## Import

You can import a datastore:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export SEL_PROJECT_ID=<selectel_project_id>
export SEL_REGION=<selectel_pool>
terraform import selectel_dbaas_redis_datastore_v1.datastore_1 <datastore_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<datastore_id>` — Unique identifier of the datastore, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the datastore ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ copy the ID under the cluster name.