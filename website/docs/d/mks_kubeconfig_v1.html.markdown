---
layout: "selectel"
page_title: "Selectel: selectel_mks_kubeconfig_v1"
sidebar_current: "docs-selectel-datasource-mks-kubeconfig-v1"
description: |-
  Provides a kubeconfig file and its fields for a Selectel Managed Kubernetes cluster.
---

# selectel\_mks\_kubeconfig_v1

Provides a kubeconfig file and its fields for a Managed Kubernetes cluster. For more information about Managed Kubernetes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-kubernetes/).

## Example Usage

### Output kubeconfig

```hcl
data "selectel_mks_kubeconfig_v1" "kubeconfig" {
  cluster_id = selectel_mks_cluster_v1.cluster_1.id
  project_id = selectel_mks_cluster_v1.cluster_1.project_id
  region     = selectel_mks_cluster_v1.cluster_1.region
}

output "kubeconfig" {
  value = data.selectel_mks_kubeconfig_v1.kubeconfig.raw_config
}
```

### Using a Kubernetes provider

```hcl
data "selectel_mks_kubeconfig_v1" "kubeconfig" {
  cluster_id = selectel_mks_cluster_v1.cluster_1.id
  project_id = selectel_mks_cluster_v1.cluster_1.project_id
  region     = selectel_mks_cluster_v1.cluster_1.region
}

provider "kubernetes" {
  host                   = data.selectel_mks_kubeconfig_v1.kubeconfig.server
  client_certificate     = base64decode(data.selectel_mks_kubeconfig_v1.kubeconfig.client_cert)
  client_key             = base64decode(data.selectel_mks_kubeconfig_v1.kubeconfig.client_key)
  cluster_ca_certificate = base64decode(data.selectel_mks_kubeconfig_v1.kubeconfig.cluster_ca_cert)
}

output "kubeconfig" {
  value = data.selectel_mks_kubeconfig_v1.kubeconfig.raw_config
}
```

## Argument Reference

* `cluster_id` - (Required) Unique identifier of the cluster.

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the cluster is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-kubernetes).

## Attributes Reference

* `raw_config` - Raw content of a kubeconfig file.

* `server` - IP address and port for a Kube API server.

* `cluster_ca_cert` - CA certificate of the cluster.

* `client_key` - Client key for authorization.

* `client_cert` - Client certificate for authorization.
