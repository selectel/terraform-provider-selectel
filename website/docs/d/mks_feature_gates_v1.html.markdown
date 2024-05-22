---
layout: "selectel"
page_title: "Selectel: selectel_mks_feature_gates_v1"
sidebar_current: "docs-selectel-datasource-mks-feature-gates-v1"
description: |-
  Provides a list of feature gates available in Selectel Managed Kubernetes.
---

# selectel\_mks\_feature_gates_v1

Provides a list of available feature gates. For more information about feature gates in Managed Kubernetes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/feature-gates/).

## Example Usage

```hcl
data "selectel_mks_feature_gates_v1" "fg" {
  project_id = selectel_vpc_project_v2.project_1.id
  region = "ru-3"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the cluster is located, for example, `ru-3`.

* `filter` - (Optional) Values to filter available feature gates:

  * `kube_version` - (Optional) Kubernetes version for which you get available feature gates.

## Attributes Reference

* `feature_gates` - List of available feature gates.

  * `kube_version` - Kubernetes version.

  * `names` - Names of the feature gates available for the specified version.