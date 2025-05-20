---
layout: "selectel"
page_title: "Selectel: selectel_mks_cluster_v1"
sidebar_current: "docs-selectel-resource-mks-cluster-v1"
description: |-
  Creates and manages a cluster in Selectel Managed Kubernetes using public API v1.
---

# selectel\_mks\_cluster\_v1

Creates and manages a Managed Kubernetes cluster using public API v1. For more information about Managed Kubernetes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-kubernetes/).

## Example usage

### High availability cluster

```hcl
resource "selectel_mks_cluster_v1" "ha_cluster" {
  name         = "cluster-1"
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-7"
  kube_version = data.selectel_mks_kube_versions_v1.versions.latest_version
}
```

### Basic cluster

```hcl
resource "selectel_mks_cluster_v1" "basic_cluster" {
  name                              = "cluster-1"
  project_id                        = selectel_vpc_project_v2.project_1.id
  region                            = "ru-7"
  kube_version                      = data.selectel_mks_kube_versions_v1.versions.latest_version
  zonal                             = true
  enable_patch_version_auto_upgrade = false
}
```

## Argument Reference

* `name` - (Required) Cluster name. Changing this creates a new cluster. The cluster name is included into the names of the cluster entities: node groups, nodes, load balancers, networks, and volumes.

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new cluster. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/cloud/managed-kubernetes/about/projects/).

* `region` - (Required) Pool where the cluster is located, for example, `ru-7`. Changing this creates a new cluster. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-kubernetes).

* `kube_version` - (Required) Kubernetes version of the cluster. Changing this upgrades the cluster version. You can retrieve information about the Kubernetes versions with the [selectel_mks_kube_versions_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/mks_kube_versions_v1) data source.

  To upgrade a patch version, the desired version should match the latest available patch version for the current minor release.

  To upgrade a minor version, the desired version should match the next available minor release with the latest patch version.

* `zonal` - (Optional) Specifies a cluster type. Changing this creates a new cluster.

  Boolean flag:

  * `false` (default) - for a high availability cluster with three master nodes located on different hosts in one pool segment.

  * `true` - for a basic cluster with one master node. Set `enable_patch_version_auto_upgrade` to `false`.

  Learn more about [Cluster types](https://docs.selectel.ru/en/cloud/managed-kubernetes/about/about-managed-kubernetes/#cluster-types).

* `enable_autorepair` - (Optional) Enables or disables node auto-repairing (worker nodes are automatically restarted). Auto-repairing is not available if you have one worker node. After auto-repairing, all data on the boot volumes are deleted. Boolean flag, the default value is `true`. Learn more about [Nodes auto-repairing](https://docs.selectel.ru/en/cloud/managed-kubernetes/node-groups/reinstall-nodes/).

* `enable_patch_version_auto_upgrade` - (Optional) Enables or disables auto-upgrading of the cluster to the latest available Kubernetes patch version during the maintenance window. Boolean flag, the default value is `true`. Must be set to false for basic clusters (if `zonal` is `true`).  Learn more about [Patch versions auto-upgrading](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/upgrade-version/).

* `network_id` - (Optional) Unique identifier of the associated OpenStack network. Changing this creates a new cluster. Learn more about the [openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_network_v2) resource in the official OpenStack documentation.

* `subnet_id` - (Optional) Unique identifier of the associated OpenStack subnet. Changing this creates a new cluster. Learn more about the [openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_subnet_v2) resource in the official OpenStack documentation.

* `maintenance_window_start` - (Optional) Time in UTC when maintenance in the cluster starts. The format is `hh:mm:ss`. Learn more about the [Maintenance window](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/set-up-maintenance-window/).

* `feature_gates` - (Optional) Enables or disables feature gates for the cluster. You can retrieve the list of available feature gates with the [selectel_mks_feature_gates_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/mks_feature_gates_v1) data source. Learn more about [Feature gates](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/feature-gates/).

* `admission_controllers` - (Optional) Enables or disables admission controllers for the cluster. You can retrieve the list of available admission controllers with the [selectel_mks_admission_controllers_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/mks_admission_controllers_v1) data source. Learn more about [Admission controllers](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/admission-controllers/).

* `private_kube_api` - (Optional) Specifies if Kube API is available from the Internet. Changing this creates a new cluster.

  Boolean flag:

  * `false` (default) - Kube API is available from the Internet;

  * `true` - Kube API is available only from the cluster network.

* `enable_audit_logs` - (Optional) Enables or disables collection of audit logs. Learn how to [configure export of audit logs to a logging system](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/logs/#configure-export-of-audit-logs).

  Boolean flag:

  * `false` (default) - Audit logs are not collected and are not available for export;

  * `true` - Audit logs are collected and available for export.

* `oidc` - (Optional) Connects an OpenID Connect (OIDC) provider to the cluster. Learn how to [configure the OIDC provider in the cluster](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/access-to-cluster-with-oidc-provider/#configure-oidc-connection).

  The block supports the following arguments:

  * `enabled` - (Required) Enables or disables authentication with OpenID Connect in the cluster.

  * `provider_name` - (Required) Name of the connection that you create. The name is only for identification purposes.

  * `issuer_url` - (Required) URL of the OIDC provider used to authenticate users who request access to the cluster. The link must start with `https://`.

  * `client_id` - (Required) Service identifier issued by the OIDC provider and used in authentication requests to the resources.

  * `username_claim` - (Optional) JWT claim to use as the username. The default value is `sub`. The content of the claim must be a unique identifier of the end user.

  * `groups_claim` - (Optional) JWT claim to use as the user's group. The default value is `groups`.

  * `ca_certs` - (Optional) Certificate in PEM format for the CA that signed your identity provider's web certificate. Optional if the certificate is issued by the public CA that Ubuntu by default considers trustworthy. Learn more about [Access to the cluster through an OIDC provider](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/access-to-cluster-with-oidc-provider/).

## Attributes Reference

* `maintenance_window_end` - Time in UTC when maintenance in the cluster ends. The format is `hh:mm:ss`. Learn more about the [Maintenance window](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/set-up-maintenance-window/).

* `kube_api_ip` - IP address of the Kube API.

* `status` - Cluster status.

## Import

You can import a cluster:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
export INFRA_REGION=<selectel_pool>
terraform import selectel_mks_cluster_v1.cluster_name <cluster_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/mks), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/cloud/managed-kubernetes/about/projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-7`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/mks/), go to **Cloud Platform** ⟶ **Kubernetes**. The pool is in the **Pool** column.

* `<cluster_id>` — Unique identifier of the cluster, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the cluster ID, in the [Control panel](https://my.selectel.ru/vpc/mks/), go to **Cloud Platform** ⟶ **Kubernetes** ⟶ the cluster page ⟶ copy the ID at the top of the page under the cluster name, near the region and pool.
