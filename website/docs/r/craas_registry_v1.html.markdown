---
layout: "selectel"
page_title: "Selectel: selectel_craas_registry_v1"
sidebar_current: "docs-selectel-resource-craas-registry-v1"
description: |-
Manages a V1 registry resource within Selectel Container Registry Service.
---

# selectel\_craas\_registry\_v1

Creates and manages  a registry in Container Registry using  public API v1. For more information about Container Registry, see the [official Selectel documentation](https://docs.selectel.ru/cloud/craas/).


## Example usage

```hcl
resource "selectel_craas_registry_v1" "registry_1" {
  name       = "my-first-registry"
  project_id = selectel_vpc_project_v2.project_1.id
}
```

## Argument Reference

* `name` - (Required) Registry name. Changing this creates a new registry. The name can contain lowercase latin characters, digits, and hyphens. The name starts with a letter and ends with a letter or a digit. It cannot exceed 20 symbols. Learn more about [Registries in Container Registry](https://docs.selectel.ru/cloud/craas/registry/).

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new registry. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

## Attributes Reference

* `status` - Registry status.

* `endpoint` - Registry endpoint. For example, `cr.selcloud.ru/my-registry`

## Import

You can import a registry:

```shell
terraform import selectel_craas_registry_v1.registry_1 <registry_id>
```

where `<registry_id>` is a unique identifier of the registry, for example, `939506d6-7621-4581-b673-eacf3db30f5b`. To get the registry ID, use [Selectel Cloud Management API](https://developers.selectel.ru/docs/selectel-cloud-platform/craas_api/).

### Environment Variables

For import, you must set environment variables:

* `SEL_TOKEN=<selectel_api_token>`
* `SEL_PROJECT_ID=<selectel_project_id>`

where:

* `<selectel_api_token>` — Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶   **API keys**  ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).
* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to Cloud Platform ⟶ project name ⟶  copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-kubernetes/about/projects/).