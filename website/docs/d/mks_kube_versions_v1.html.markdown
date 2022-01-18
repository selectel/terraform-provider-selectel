---
layout: "selectel"
page_title: "Selectel: selectel_mks_kube_versions_v1"
sidebar_current: "docs-selectel-datasource-mks-kube-versions-v1"
description: |-
Get all supported kube versions for a Selectel Managed Kubernetes cluster.
---

# selectel\_mks\_kube_versions_v1

Use this data source to get all supported kube versions for a Managed Kubernetes cluster.

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
}

data "selectel_mks_kube_versions_v1" "versions" {
  project_id = "${selectel_vpc_project_v2.project_1.id}"
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

The following arguments are supported:

* `project_id` - (Required) An associated Selectel VPC project.

* `region` - (Required) A Selectel VPC region.

## Attributes Reference

The following attributes are exported:

* `latest_version` - The most recent version available.

* `default_version` - The currently supported version that is suggested to be used by default.

* `versions` - The list of all supported versions.
