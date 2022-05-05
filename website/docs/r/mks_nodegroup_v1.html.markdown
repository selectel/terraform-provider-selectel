---
layout: "selectel"
page_title: "Selectel: selectel_mks_nodegroup_v1"
sidebar_current: "docs-selectel-resource-mks-nodegroup-v1"
description: |-
  Manages a V1 nodegroup resource within Selectel Managed Kubernetes Service.
---

# selectel\_mks\_nodegroup\_v1

Manages a V1 nodegroup resource within Selectel Managed Kubernetes Service.

## Example usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
}

resource "selectel_mks_cluster_v1" "cluster_1" {
  name         = "cluster-1"
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
  kube_version = "1.16.8"
}

resource "selectel_mks_nodegroup_v1" "nodegroup_1" {
  cluster_id        = "${selectel_mks_cluster_v1.cluster_1.id}"
  project_id        = "${selectel_mks_cluster_v1.cluster_1.project_id}"
  region            = "${selectel_mks_cluster_v1.cluster_1.region}"
  availability_zone = "ru-3a"
  nodes_count       = 3
  cpus              = 2
  ram_mb            = 2048
  volume_gb         = 20
  volume_type       = "fast.ru-3a"
  labels            = {
    "label-key0": "label-value0",
    "label-key1": "label-value1",
    "label-key2": "label-value2",
  }
  taints {
    key = "test-key-0"
    value = "test-value-0"
    effect = "NoSchedule"
  }
  taints {
    key = "test-key-1"
    value = "test-value-1"
    effect = "NoExecute"
  }
  taints {
    key = "test-key-2"
    value = "test-value-2"
    effect = "PreferNoSchedule"
  }
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) An associated MKS cluster.
  Changing this creates a new nodegroup.

* `project_id` - (Required) An associated Selectel VPC project.
  Changing this creates a new nodegroup.

* `region` - (Required) A Selectel VPC region of where the nodegroup is located.
  Changing this creates a new nodegroup.

* `availability_zone` (Required) An OpenStack availability zone for all nodes in the nodegroup.
  Changing this creates a new nodegroup.

* `nodes_count` (Required) Count of worker nodes in the nodegroup.
  Changing this resizes the nodegroup according to the new nodes count.
  As long as `enable_autoscale` is set to true, changing this will not affect the size of the nodegroup.

* `keypair_name` (Optional) Name of the SSH key that will be added to all nodes.
  Changing this creates a new nodegroup.

* `affinity_policy` (Optional) An argument to tune nodes affinity policy.
  Changing this creates a new nodegroup.

* `cpus` (Optional) CPU count for each node. It can be omitted only in cases when `flavor_id` is set.
  Changing this creates a new nodegroup.

* `ram_mb` (Optional) RAM count in MB for each node. It can be omitted only in cases when `flavor_id` is set.
  Changing this creates a new nodegroup.

* `volume_gb` (Optional) Volume size in GB for each node. It can be omitted only in cases
   when `flavor_id` is set and `local_volume` is true.
   Changing this creates a new nodegroup.

* `volume_type` (Optional) An OpenStack blockstorage volume type for each node. It can be omitted only in cases
   when `flavor_id` is set and `local_volume` is true.
   Changing this creates a new nodegroup.

* `local_volume` (Optional) Represents if nodes will use local volume.
  Accepts true or false. Defaults to false.
  Changing this creates a new nodegroup.

* `flavor_id` (Optional) An OpenStack flavor identifier for all nodes in the nodegroup. It can be omitted in most cases.
  Changing this creates a new nodegroup.

* `labels` (Optional) Represents a map containing a set of Kubernetes labels that will be applied
  for each node in the group. The keys must be user-defined.

* `taints` (Optional) Represents a list of Kubernetes taints that will be applied for each node in the group.

* `enable_autoscale` (Optional) Specifies if a nodegroup autoscaling option has to be turned on/off.
  Accepts true or false. Default is false.
  `autoscale_min_nodes` and `autoscale_max_nodes` must be specified.

* `autoscale_min_nodes` (Optional) Represents a minimum possible number of worker nodes in the nodegroup.

* `autoscale_max_nodes` (Optional) Represents a maximum possible number of worker nodes in the nodegroup.

## Attributes Reference

The following attributes are exported:

* `nodes` - Contains a list of all nodes in the nodegroup.

## Import

Nodegroup can be imported using a combined ID using the following format: ``<cluster_id>/<nodegroup_id>``

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN SEL_PROJECT_ID=SELECTEL_VPC_PROJECT_ID SEL_REGION=SELECTEL_VPC_REGION terraform import selectel_mks_nodegroup_v1.nodegroup_1 b311ce58-2658-46b5-b733-7a0f418703f2/63ed5342-b22c-4c7a-9d41-c1fe4a142c13
```