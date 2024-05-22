---
layout: "selectel"
page_title: "Selectel: selectel_mks_admission_controllers_v1"
sidebar_current: "docs-selectel-datasource-mks-admission-controllers-v1"
description: |-
  Provides a list of admission controllers available in Selectel Managed Kubernetes.
---

# selectel\_mks\_admission_controllers_v1

Provides a list of available admission controllers. For more information about admission controllers in Managed Kubernetes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/admission-controllers/).

## Example Usage

```hcl
data "selectel_mks_admission_controllers_v1" "admission_controllers_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region = "ru-3"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the cluster is located, for example, `ru-3`.

* `filter` - (Optional) Values to filter available admission controllers:

  * `kube_version` - (Optional) Kubernetes version for which you get available admission controllers.

## Attributes Reference

* `admission_controllers` - List of available admission controllers.

  * `kube_version` - Kubernetes version.

  * `names` - Names of the admission controllers available for the specified Kubernetes version.