---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_clickhouse_datastore_v2"
sidebar_current: "docs-selectel-resource-dbaas-clickhouse-datastore-v2"
description: |-
  Creates and manages a ClickHouse cluster in Selectel Managed Databases using public API v2.
---

# selectel\_dbaas\_clickhouse\_datastore\_v2

Creates and manages a ClickHouse cluster using public API v2. For more information about Managed Databases, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-databases/clickhouse/).

## Example usage

```hcl
resource "selectel_dbaas_clickhouse_datastore_v2" "cluster_1" {
  name       = "cluster-1"
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  type_id    = data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id
  subnet_id  = selectel_vpc_subnet_v2.subnet.subnet_id
  password   = "secretsecretsecretsecret"

  node_group {
    name       = "keepers"
    role       = "KEEPER"
    node_count = 3
    flavor {
      id   = data.selectel_dbaas_flavor_v2.keeper_flavor.flavors[0].id
      type = "FIXED"
    }
  }

  node_group {
    name           = "shard1"
    role           = "DATA"
    node_count     = 2
    weight         = 100
    has_public_ips = true
    flavor {
      vcpus     = 2
      ram       = 8192
      disk      = 32
      disk_type = "NETWORK_ULTRA"
      type      = "FLEXIBLE"
    }
  }

  node_group {
    name           = "shard2"
    role           = "DATA"
    node_count     = 1
    weight         = 50
    has_public_ips = true
    flavor {
      vcpus     = 2
      ram       = 8192
      disk      = 32
      disk_type = "NETWORK_ULTRA"
      type      = "FLEXIBLE"
    }
  }

  security_groups = ["796f1f0a-d97d-4a8e-904e-4fd5ef57465c", "b9c2e73d-a6c5-4def-994d-ce85e3ce98d3"]
}
```

## Argument Reference

* `name` - (Required) Cluster name. Changing this creates a new cluster.

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new cluster. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the cluster is located, for example, `ru-3`. Changing this creates a new cluster. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `subnet_id` - (Required) Unique identifier of the associated subnet. Changing this creates a new cluster. Retrieved from the [selectel_vpc_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_subnet_v2) resource for a public subnet, or from the [openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_subnet_v2) resource of the OpenStack provider for a private subnet.

* `type_id` - (Required) Unique identifier of the cluster type. Changing this creates a new cluster. Retrieved from the [selectel_dbaas_datastore_type_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_datastore_type_v2) data source.

* `password` - (Required) Password for the cluster. Changing this updates the password.

* `node_group` - (Required) List of node groups in the cluster. A cluster must contain at least one node group.

  * `name` - (Required) Name of the node group. Must be unique within the cluster.

  * `role` - (Required) Role of the node group. Available values are `DATA` and `KEEPER`. The `KEEPER` role is used for ClickHouse Keeper nodes.

  * `node_count` - (Required) Number of nodes in the group. Must be at least `1` for `DATA` role and `3` for `KEEPER` role.

  * `weight` - (Optional) Weight of the node group. Used for `DATA` role groups to distribute data across shards. Must be greater than `0` for `DATA` role. Not applicable for `KEEPER` role. The default value is `0`.

  * `has_public_ips` - (Optional) Assigns public IP addresses to the nodes in the group. The network configuration must meet the requirements. Not applicable for `KEEPER` role. Learn more about [public IP addresses and the required network configuration](https://docs.selectel.ru/en/cloud/managed-databases/clickhouse/public-ip/).

  * `flavor` - (Required) Flavor configuration for the node group.

    * `id` - (Optional) Unique identifier of the predefined flavor. Required for `FIXED` flavor type. Learn more about available flavors for [ClickHouse](https://docs.selectel.ru/en/cloud/managed-databases/clickhouse/configurations/).

    * `type` - (Required) Flavor type. Available values are `FIXED` and `FLEXIBLE`. `FIXED` type uses predefined flavors, `FLEXIBLE` allows you to set custom vCPUs, RAM, and disk.

    * `vcpus` - (Optional) Number of vCPUs. Required for `FLEXIBLE` flavor type.

    * `ram` - (Optional) Amount of RAM in MB. Required for `FLEXIBLE` flavor type.

    * `disk` - (Optional) Volume size in GB. Required for `FLEXIBLE` flavor type.

    * `disk_type` - (Optional) Volume type. Available values are `LOCAL` and `NETWORK_ULTRA`. Required for `FLEXIBLE` flavor type. Learn more about volumes for [ClickHouse](https://docs.selectel.ru/en/cloud/managed-databases/clickhouse/volumes/).

* `config` - (Optional) Configuration parameters for the cluster. You can retrieve information about available configuration parameters with the [selectel_dbaas_clickhouse_configuration_parameter_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_clickhouse_configuration_parameter_v2) data source. Setting a parameter to `null` resets it to the default value.

* `security_groups` - (Optional) List of security groups. If no security group UUIDs are specified when creating the cluster, a default security group will be created and its UUID will be assigned automatically. A cluster must have at least one security group. Learn more about security groups for [ClickHouse](https://docs.selectel.ru/en/cloud/managed-databases/clickhouse/network-access-control/#security-groups-in-managed-databases).

* `log_platform` - (Optional) Name of an existing or a new log group in the [Logs](https://docs.selectel.ru/en/logs/about-logs/) service. The name must start with the prefix `s/dbaas/`. It can contain uppercase and lowercase letters, digits and symbols (underscore, hyphen, forward slash, period and hash). The name cannot exceed 512 symbols. For example, `s/dbaas/My-first-group`. Learn more about [Logs](https://docs.selectel.ru/en/managed-databases/clickhouse/logs/).

  * `log_group` - (Required) Name of the log group.

* `allow_reduce_nodes` - (Optional) Allows forced reduction of node count in node groups. The default value is `false`.

## Important notes about node groups

* The order of `node_group` blocks is significant because `node_group` is a list (`TypeList`). After adding or removing a group in the middle, Terraform shows a difference for subsequent groups. This is only a display artifact — the provider manages groups by name.

* Do not change the `name` of an existing node group. The provider identifies groups by `name`. Keeping the name allows in-place updates (`node_count`, `weight`, `has_public_ips`, `flavor`). Changing a name when the number of groups is unchanged causes a `terraform plan` error. Changing a name while also adding or removing groups is treated as deleting the old group and creating a new one.

* Do not change the `role` of an existing node group. Changing the role is prohibited by the provider and returns an error: `node_group: changing role of node group "<name>" is not allowed`.

* You can freely change `node_count`, `weight`, `has_public_ips`, and `flavor` for existing groups.

* Add new node groups to the end of the list. This minimizes display differences in the Terraform state.

* When deleting a node group from the middle or beginning of the list, Terraform shows a difference for the groups after the deleted one. The correct group is deleted by name. This difference in display is expected and does not affect the actual infrastructure.

## Attributes Reference

* `status` - Cluster status.

* `state` - Cluster state.

* `node_group` - List of node groups. Each group includes the following computed attributes in addition to the configured ones:

  * `status` - Status of the node group.

  * `instances` - List of instances in the node group.

    * `id` - Unique identifier of the instance.

    * `ip` - IP address of the instance.

    * `floating_ip` - Public IP address of the instance.

    * `availability_zone` - Availability zone of the instance.

    * `hostname` - Hostname of the instance.

## Import

You can import a cluster:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
export INFRA_REGION=<selectel_pool>
terraform import selectel_dbaas_clickhouse_datastore_v2.cluster_1 <datastore_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<datastore_id>` — Unique identifier of the cluster, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the cluster ID in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ copy the ID under the cluster name.
