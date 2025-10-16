---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_mysql_datastore_v1"
sidebar_current: "docs-selectel-resource-dbaas-mysql-datastore-v1"
description: |-
  Creates and manages a MySQL cluster in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_mysql\_datastore\_v1

Creates and manages a MySQL cluster using public API v1. Applicable to MySQL sync and MySQL semi-sync clusters. For more information about Managed Databases, see the official Selectel documentation for [MySQL sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-sync/) and [MySQL semi-sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-semi-sync/).

## Example usage

```hcl
resource "selectel_dbaas_mysql_datastore_v1" "cluster_1" {
  name       = "cluster-1"
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  type_id    = data.selectel_dbaas_datastore_type_v1.datastore_type_1.datastore_types[0].id
  subnet_id  = selectel_vpc_subnet_v2.subnet.subnet_id
  node_count = 3
  flavor {
    vcpus     = 4
    ram       = 4096
    disk      = 32
    disk_type = "network-ultra"
  }
  security_groups = ["796f1f0a-d97d-4a8e-904e-4fd5ef57465c", "b9c2e73d-a6c5-4def-994d-ce85e3ce98d3"]
}
```

## Argument Reference

* `name` - (Required) Cluster name. Changing this creates a new cluster.

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new cluster. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new cluster. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `subnet_id` - (Required) Unique identifier of the associated OpenStack network. Changing this creates a new cluster. Learn more about the [openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_network_v2) resource in the official OpenStack documentation.

* `type_id` - (Required) Unique identifier of the cluster type. Changing this creates a new cluster. Retrieved from the [selectel_dbaas_datastore_type_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_datastore_type_v1) data source.

* `node_count` - (Required) Number of nodes in the cluster. The available range for MySQL semi-sync is from 1 to 3. Available values for MySQL sync are `1` and `3`. Learn more about [Replication](https://docs.selectel.ru/en/cloud/managed-databases/about/about-managed-databases/#fault-tolerance-and-replication).

* `flavor_id` - (Optional) Unique identifier of the flavor for the cluster. Can be skipped when `flavor` is set. You can retrieve information about available flavors with the [selectel_dbaas_flavor_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_flavor_v1) data source.

* `flavor` - (Optional) Flavor configuration for the cluster. You can retrieve information about available flavors with the [selectel_dbaas_flavor_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_flavor_v1) data source. Learn more about available configurations for [MySQL sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-sync/configurations/) and [MySQL semi-sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-semi-sync/configurations/).

  * `vcpus` - (Required) Number of vCPUs.

  * `ram` - (Required) Amount of RAM in MB.

  * `disk` - (Required) Volume size in GB.

  * `disk_type` - (Optional) Volume type. Available values are `local` and `network-ultra`. The default value is `local.` Learn more about volumes for [MySQL sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-sync/volumes/) and [MySQL semi-sync](https://docs.selectel.ru/en/cloud/managed-databases/mysql-semi-sync/volumes/).

* `firewall` - (Deprecated) Remove this argument as it is no longer in use and will be removed in the next major version of the provider. To manage a list of IP-addresses with access to the cluster, use the [selectel_dbaas_firewall_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_firewall_v1) resource.

* `restore` - (Optional) Restores parameters for the cluster. Changing this creates a new cluster.

  * `datastore_id` - (Optional) Unique identifier of the cluster from which you restore. To get the cluster ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ copy the ID under the cluster name.

  * `target_time` - (Optional) Time within seven previous days when you have the cluster state to restore.

* `config` - (Optional) Configuration parameters for the cluster. You can retrieve information about available configuration parameters with the [selectel_dbaas_configuration_parameter_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_configuration_parameter_v1) data source.

* `floating_ips` - (Optional) Assigns public IP addresses to the nodes in the cluster. The network configuration must meet the requirements. Learn more about [public IP addresses and the required network configuration](https://docs.selectel.ru/en/cloud/managed-databases/mysql-sync/public-ip/).

  * master - (Required) Number of public IPs associated with the master. Available values are `0` and `1`.

  * replica - (Required) Number of public IPs associated with the replicas. The minimum value is `0`. The maximum value must be 1 less than the value of the `node_count` argument.

* `backup_retention_days` - (Optional) Number of days to retain backups.

* `logs` - (Optional) Name of an existing or a new log group in the [Logs](https://docs.selectel.ru/en/logs/about-logs/) service. The name must start with the prefix 's/dbaas/'. It can contain uppercase and lowercase letters, digits and symbols (underscore, hyphen, forward slash, period and hash). The name cannot exceed 512 symbols. For example, s/dbaas/My-first-group. Learn more about logs for [MySQL sync](https://docs.selectel.ru/en/managed-databases/mysql-sync/logs/) and [MySQL semi-sync](https://docs.selectel.ru/en/managed-databases/mysql-semi-sync/logs/).

* `security_groups` - (Optional) List of security groups. If no security group UUIDs are specified when creating the datastore, a default security group will be created and its UUID will be assigned automatically. A datastore must have at least one security group. Learn more about security groups for [MySQL sync](https://docs.selectel.ru/en/managed-databases/mysql-sync/network-access-control/#security-groups-in-managed-databases) and [MySQL semi-sync](https://docs.selectel.ru/en/managed-databases/mysql-semi-sync/network-access-control/#security-groups-in-managed-databases).

## Attributes Reference

* `status` - Cluster status.

* `connections` - DNS addresses to connect to the cluster.

## Import

You can import a cluster:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
export INFRA_REGION=<selectel_pool>
terraform import selectel_dbaas_mysql_datastore_v1.datastore_1 <datastore_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<datastore_id>` — Unique identifier of the cluster, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the cluster ID in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ copy the ID under the cluster name.
