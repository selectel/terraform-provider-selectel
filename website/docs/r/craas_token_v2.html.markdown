---
layout: "selectel"
page_title: "Selectel: selectel_craas_token_v2"
sidebar_current: "docs-selectel-resource-craas-token-v2"
description: |-
  Creates and manages tokens in Selectel Container Registry using public API v2.
---

# selectel\_craas\_token\_v2

Creates and manages tokens in Container Registry using public API v2. For more information about Container Registry, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/craas/).

## Basic usage example

```hcl
resource "selectel_craas_token_v2" "token_1" {
  project_id     = selectel_vpc_project_v2.project_1.id
  name           = "token-name"
  mode_rw        = true
  all_registries = true
  registry_ids   = []
  is_set         = true
  expires_at     = "2029-01-01T00:00:00Z"
}
```

## Docker CLI login example

```hcl
resource "selectel_craas_token_v2" "token_1" {
  project_id     = selectel_vpc_project_v2.project_1.id
  name           = "terraform-token-270295000"
  mode_rw        = true
  all_registries = true
  registry_ids   = []
  is_set         = true
  expires_at     = "2029-01-01T00:00:00Z"
}

output "registry_username" {
  value     = selectel_craas_token_v2.token_1.username
  sensitive = true
}

output "registry_token" {
  value     = selectel_craas_token_v2.token_1.token
  sensitive = true
}
```

```shell
REGISTRY_USERNAME=$(terraform output -raw registry_username)
REGISTRY_TOKEN=$(terraform output -raw registry_token)
echo $REGISTRY_TOKEN | docker login cr.selcloud.ru --username $REGISTRY_USERNAME --password-stdin
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new token. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `mode_rw` - (Required) Specifies the scope of access to registries. Changing this updates the token.

  Boolean flag:

  * `false` (default) — for read-only scope of access.
  * `true` — for read and write scope of access.

* `is_set` - (Required) Specifies if the token lifetime is limited. Changing this updates the token.

  Boolean flag:

  * `false` (default) — for an unlimited token lifetime.
  * `true` — for a limited token lifetime. Requires the `expires_at` argument.

* `expires_at` - (Optional) Token lifetime in the RFC3339 timestamp format, for example, `2025-03-09T12:58:49Z`. Changing this updates the token. Required when `is_set` is `true`.

* `all_registries` - (Required) Specifies if the token provides access to all registries. Changing this updates the token.

  Boolean flag:

  * `false` (default) — for access to the specific registry. Requires the `registry_ids` argument.
  * `true` — for access to all registries. The token will be applicable to all new registries that you will create in the project.

* `registry_ids` - (Optional) Unique identifier of the specific registry access to which is granted. Changing this updates the token. Required when `all _registries` is `false`. To get the registry ID, in the [Control panel](https://my.selectel.ru/vpc/default/craas/), go to **Products** ⟶ **Container Registry** ⟶ copy the ID under the registry name.

* `name` - (Optional) Token name. Changing this updates the token.

## Attributes Reference

* `username` - (Sensitive) Username to access Container Registry.

* `token` - (Sensitive) Token to access Container Registry.
