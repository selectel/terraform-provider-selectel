---
layout: "selectel"
page_title: "Selectel: selectel_mks_kube_versions_v1"
sidebar_current: "docs-selectel-datasource-mks-kube-versions-v1"
description: |-
  Provides a list of Kubernetes versions supported in a Selectel Managed Kubernetes cluster.
---

# selectel\_mks\_kube_versions_v1

Provides a list of supported Kubernetes versions for a Managed Kubernetes cluster. For more information about Managed Kubernetes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-kubernetes/about/about-managed-kubernetes/).

## Example Usage

```hcl
data "selectel_mks_kube_versions_v1" "versions" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}

output "latest_version" {
  value = data.selectel_mks_kube_versions_v1.versions.latest_version
}

output "default_version" {
  value = data.selectel_mks_kube_versions_v1.versions.default_version
}

output "versions" {
  value = data.selectel_mks_kube_versions_v1.versions.versions
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the cluster is located, for example, `ru-3`.

## Attributes Reference

* `latest_version` - The most recent version.

* `default_version` - Kubernetes version used by default.

* `versions` - List of the supported versions.