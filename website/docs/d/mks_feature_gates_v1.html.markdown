---
layout: "selectel"
page_title: "Selectel: selectel_mks_feature_gates_v1"
sidebar_current: "docs-selectel-datasource-mks-feature-gates-v1"
description: |-
  Get information on Selectel MKS available feature gates.
---

# selectel\_mks\_feature_gates_v1

Use this data source to get available feature-gates within Selectel MKS API Service.

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
}

data "selectel_mks_feature_gates_v1" "fg" {
  project_id = "${selectel_vpc_project_v2.project_1.id}"
  region = "ru-3"
  filter {
    kube_version = "1.22.2"
  }
}
```

## Argument Reference

The following arguments are supported

* `project_id` - (Required) An associated Selectel VPC project.

* `region` - (Required) A Selectel VPC region.

* `filter` - (Optional) One or more values used to look up available feature gates.

**filter**

- `kube_version` - (Optional) Kubernetes version to look up the available feature gates.

## Attributes Reference

The following attributes are exported:

* `feature_gates` - Contains a list of the found available feature gates.

**feature_gates**

- `kube_version_minor` - Kubernetes version.
- `names` - Names of the feature gates available for the specified version.
