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
  project_id       = selectel_vpc_project_v2.project_1.id
  name             = "token-name"
  mode_rw          = true
  all_registries   = true
  registry_ids     = []
  is_set           = true
  expires_at       = "2029-01-01T00:00:00Z"
}
```

## Docker CLI login example

```hcl
resource "selectel_craas_token_v2" "token_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  name              = "terraform-token-270295000"
  mode_rw           = true
  all_registries    = true
  registry_ids      = []
  is_set            = true
  expires_at        = "2029-01-01T00:00:00Z"
}

output "registry_username" {
  value = selectel_craas_token_v2.token_1.username
  sensitive = true
}

output "registry_token" {
  value = selectel_craas_token_v2.token_1.token
  sensitive = true
}
```

```shell
REGISTRY_USERNAME=$(terraform output -raw registry_username)
REGISTRY_TOKEN=$(terraform output -raw registry_token)
echo $REGISTRY_TOKEN | docker login cr.selcloud.ru --username $REGISTRY_USERNAME --password-stdin
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.
  Changing this creates a new token. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `mode_rw` - (Required) Enable scope access read-write mode.
  Changing this update token.

* `is_set` - (Required) Token unlimited lifetime. Use false if you want unlimited lifetime token. False ignore *expires_at* option.
  Changing this update token. 

* `all_registries` - (Required) Access to all available registries.
  Changing this update token. 

* `name` - (Optional) Token name.
  Changing this update token.

* `registry_ids` - (Optional) Access to specific registries by IDs. 
  Changing this update token.

* `expires_at` - (Optional) Token lifetime. 
  Changing this update token. Use RFC 3339 date and time format.

## Attributes Reference

* `username` - (Sensitive) Username to access Container Registry.

* `token` - (Sensitive) Token to access Container Registry.
