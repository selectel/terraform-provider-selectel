---
layout: "selectel"
page_title: "Selectel: selectel_mks_kubeconfig_v1"
sidebar_current: "docs-selectel-datasource-mks-kubeconfig-v1"
description: |-
  Get kubeconfig for a Selectel Managed Kubernetes cluster.
---

# selectel\_mks\_kubeconfig_v1

Use this data source to get kubeconfig and its fields for a Managed Kubernetes cluster.

## Example Usage

```hcl
resource "selectel_mks_cluster_v1" "cluster_1" {
  name                              = var.cluster_name
  project_id                        = var.project_id
  region                            = var.region
  kube_version                      = var.kube_version
  enable_autorepair                 = var.enable_autorepair
  enable_patch_version_auto_upgrade = var.enable_patch_version_auto_upgrade
  network_id                        = var.network_id
  subnet_id                         = var.subnet_id
  maintenance_window_start          = var.maintenance_window_start
}

data "selectel_mks_kubeconfig_v1" "kubeconfig" {
  cluster_id  = selectel_mks_cluster_v1.cluster_1.id
  project_id  = var.project_id
  region      = var.region
}

provider "kubernetes" {
  host                    = data.selectel_mks_kubeconfig_v1.kubeconfig.server
  client_certificate      = data.selectel_mks_kubeconfig_v1.kubeconfig.cluster_ca_cert
  client_key              = data.selectel_mks_kubeconfig_v1.kubeconfig.client_key
  cluster_ca_certificate  = data.selectel_mks_kubeconfig_v1.kubeconfig.client_cert
}

output "kubeconfig" {
  value = data.selectel_mks_kubeconfig_v1.kubeconfig.raw_config
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) ID of the Managed Kubernetes cluster.

* `project_id` - (Required) Project ID where the cluster is placed.

* `region`     - (Required) Region where the cluster is placed.

## Attributes Reference

The following attributes are exported:

* `raw_config` - Raw content of a kubeconfig file.

* `server` - IP address and port for a kube-API server.

* `cluster_ca_cert` - K8s cluster CA certificate.

* `client_key` - Client key for authorization.

* `client_cert` - Client cert for authorization.
