---
layout: "selectel"
page_title: "Selectel: selectel_craas_registry_v1"
sidebar_current: "docs-selectel-resource-craas-registry-v1"
description: |-
  Creates and manages a registry in Selectel Container Registry using public API v1.
---

# selectel\_craas\_registry\_v1

Creates and manages a registry in Container Registry using public API v1. For more information about Container Registry, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/craas/).

## Example usage

```hcl
resource "selectel_craas_registry_v1" "registry_1" {
  name       = "my-first-registry"
  project_id = selectel_vpc_project_v2.project_1.id
}
```

## Argument Reference

* `name` - (Required) Registry name. Changing this creates a new registry. The name can contain lowercase latin characters, digits, and hyphens. The name starts with a letter and ends with a letter or a digit. It cannot exceed 20 symbols. Learn more about [Registries in Container Registry](https://docs.selectel.ru/en/cloud/craas/registry/).

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new registry. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

## Attributes Reference

* `status` - Registry status.

* `endpoint` - Registry endpoint. For example, `cr.selcloud.ru/my-registry`

## Import

You can import a registry:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
terraform import selectel_craas_registry_v1.registry_1 <registry_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/craas), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<registry_id>` — Unique identifier of the registry, for example, `939506d6-7621-4581-b673-eacf3db30f5b`. To get the registry ID, use [Selectel Cloud Management API](https://developers.selectel.ru/docs/selectel-cloud-platform/craas_api/).