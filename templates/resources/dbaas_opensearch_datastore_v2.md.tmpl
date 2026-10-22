---
page_title: "Selectel: selectel_dbaas_opensearch_datastore_v2"
description: |-
  Creates and manages an OpenSearch cluster in Selectel Managed Databases using public API v2.
---

# selectel\_dbaas\_opensearch\_datastore\_v2

Creates and manages a OpenSearch cluster using public API v2. For more information about Managed Databases, see the [official Selectel documentation](https://docs.selectel.ru/en/managed-databases/opensearch/).

## Example usage

```terraform
resource "selectel_dbaas_opensearch_datastore_v2" "cluster_1" {
  name       = "cluster-1"
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  type_id    = data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id
  subnet_id  = selectel_vpc_subnet_v2.subnet.subnet_id
  password   = "secretsecretsecretsecret"

  node_group {
    key        = "managers"
    name       = "managers"
    role       = "MANAGER"
    node_count = 3
    flavor {
      id   = data.selectel_dbaas_flavor_v2.manager_flavor.flavors[0].id
      type = "FIXED"
    }
  }

  node_group {
    key            = "data1"
    name           = "data1"
    role           = "DATA"
    node_count     = 2
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
    key            = "data2"
    name           = "data2"
    role           = "DATA"
    node_count     = 1
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

* `name` - (Required) Cluster name. Changing this updates the name.

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new cluster. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the cluster is located, for example, `ru-3`. Changing this creates a new cluster. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `subnet_id` - (Required) Unique identifier of the associated subnet. Changing this creates a new cluster. Retrieved from the [selectel_vpc_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_subnet_v2) resource for a public subnet, or from the [openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_subnet_v2) resource of the OpenStack provider for a private subnet.

* `type_id` - (Required) Unique identifier of the cluster type. Changing this creates a new cluster. Retrieved from the [selectel_dbaas_datastore_type_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_datastore_type_v2) data source.

* `password` - (Required) Password for the cluster. Changing this updates the password in the cluster.

* `node_group` - (Required) List of node groups in the cluster. A cluster must contain at least one `DATA` node group and one `MANAGER` node group. Each node group is identified by its `key`.

  * `key` - (Required) Stable identifier of the node group used by the provider to match node groups between the configuration and the state. Must be unique within the cluster and not empty. Must not change during the lifetime of the node group: changing `key` is treated as deleting the old node group and creating a new one. To rename a node group, change `name` and keep `key` unchanged.

  * `name` - (Required) Name of the node group. Can be changed in-place by keeping the `key` unchanged.

  * `role` - (Required) Role of the node group. Available values are `DATA`, `MANAGER` and `DASHBOARD`. Node group with role `DASHBOARD` is not required.

  * `node_count` - (Required) Number of nodes in the group. Must be at least `1` for `DATA` role, exactly `3` for `MANAGER` role, exactly `1` for `DASHBOARD` role. A cluster must contain exactly one `MANAGER` node group and at least one `DATA` node group.

  * `has_public_ips` - (Optional) Assigns public IP addresses to the nodes in the group. The network configuration must meet the requirements. Not applicable for `DASHBOARD` role. Learn more about [public IP addresses and the required network configuration](https://docs.selectel.ru/en/managed-databases/opensearch/public-ip/).

  * `flavor` - (Required) Flavor configuration for the node group.

    * `id` - (Optional) Unique identifier of the predefined flavor. Required for `FIXED` flavor type. Learn more about available flavors for [OpenSearch](https://docs.selectel.ru/en/managed-databases/opensearch/configurations/).

    * `type` - (Required) Flavor type. Available values are `FIXED` and `FLEXIBLE`. `FIXED` type uses predefined flavors, `FLEXIBLE` allows you to set custom vCPUs, RAM, and disk.

    * `vcpus` - (Optional) Number of vCPUs. Required for `FLEXIBLE` flavor type.

    * `ram` - (Optional) Amount of RAM in MB. Required for `FLEXIBLE` flavor type.

    * `disk` - (Optional) Volume size in GB. Required for `FLEXIBLE` flavor type.

    * `disk_type` - (Optional) Volume type. Available values are `LOCAL` and `NETWORK_ULTRA`. Required for `FLEXIBLE` flavor type. Learn more about volumes for [OpenSearch](https://docs.selectel.ru/en/managed-databases/opensearch/volumes/).

* `security_groups` - (Optional) List of security group UUIDs. If no security group UUIDs are specified when creating the cluster, the cluster is created without security groups — no default security group is assigned automatically. Changing this updates the security groups of the cluster. Learn more about security groups for [OpenSearch](https://docs.selectel.ru/en/managed-databases/opensearch/network-access-control/#security-groups-in-managed-databases).

* `log_platform` - (Optional) Name of an existing or a new log group in the [Logs](https://docs.selectel.ru/en/logs/about-logs/) service. The name must start with the prefix `s/dbaas/`. It can contain uppercase and lowercase letters, digits and symbols (underscore, hyphen, forward slash, period and hash). The name cannot exceed 512 symbols. For example, `s/dbaas/My-first-group`. Learn more about [Logs](https://docs.selectel.ru/en/managed-databases/opensearch/logs/).

  * `log_group` - (Required) Name of the log group.

## Important notes about node groups

* The provider identifies node groups by the `key` field, not by their position in the list or by `name`. `node_group` is a list (`TypeList`), so the order of blocks is reflected in the state, but it does not affect which group is created, updated or deleted.

* Do not change the `key` of an existing node group. Changing `key` is treated as deleting the old node group and creating a new one. To rename a node group, change only `name` and keep `key` unchanged — the group is renamed in-place, without recreation.

* Changing `node_count`, `has_public_ips` or `flavor` of an existing node group updates it in-place.

* Do not change the `role` of an existing node group. Changing the role is prohibited by the provider and returns an error: `node_group: changing role of node group with key "<key>" is not allowed`.

* Adding or removing a node group in the middle of the list, or reordering blocks, makes Terraform show a difference for the affected positions in the plan. This is only a display artifact — the provider matches node groups by `key`, and no redundant actions are performed. Adding new node groups to the end of the list minimizes the display churn.

* Node groups that were created outside of Terraform (for example, from the Control panel) are added to the state during the next refresh with the `imported-<node_group_id>` key. Such a node group is not described in the configuration, so the next `terraform apply` deletes it. To keep the node group, add it to the configuration with the `imported-<node_group_id>` key, or replace the key in the state with your own value first.

## Attributes Reference

* `status` - Cluster status.

* `state` - Cluster state.

* `node_group` - List of node groups. Each group includes the following computed attributes in addition to the configured ones:

  * `id` - Unique identifier of the node group.

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
terraform import selectel_dbaas_opensearch_datastore_v2.cluster_1 <datastore_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<datastore_id>` — Unique identifier of the cluster, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the cluster ID in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ copy the ID under the cluster name.
