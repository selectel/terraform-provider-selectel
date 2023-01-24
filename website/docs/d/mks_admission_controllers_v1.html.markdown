---
layout: "selectel"
page_title: "Selectel: selectel_mks_admission_controllers_v1"
sidebar_current: "docs-selectel-datasource-mks-admission-controllers-v1"
description: |-
  Get information on Selectel MKS available admission controllers.
---

# selectel\_mks\_admission_controllers_v1

Use this data source to get available admission-controllers within Selectel MKS API Service.

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
}

data "selectel_mks_admission_controllers_v1" "ac" {
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

* `filter` - (Optional) One or more values used to look up available admission controllers.

**filter**

- `kube_version` - (Optional) Kubernetes version to look up the available admission controllers.

## Attributes Reference

The following attributes are exported:

* `admission_controllers` - Contains a list of the found available admission controllers.

**admission_controllers**

- `kube_version` - Kubernetes version.
- `names` - Names of the admission controllers available for the specified version.
