---
layout: "selectel"
page_title: "Selectel: selectel_craas_token_v1"
sidebar_current: "docs-selectel-resource-craas-token-v1"
description: |-
  Creates and manages tokens in Selectel Container Registry using public API v1.
---

# selectel\_craas\_token\_v1

Creates and manages tokens in Container Registry using public API v1. For more information about Container Registry, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/craas/).

## Basic usage example

```hcl
resource "selectel_craas_token_v1" "token_1" {
  project_id = selectel_vpc_project_v2.project_1.id
}
```

## Docker CLI login example

```hcl
resource "selectel_craas_token_v1" "token_1" {
  project_id = selectel_vpc_project_v2.project_1.id
}

output "registry_username" {
  value = selectel_craas_token_v1.token_1.username
  sensitive = true
}

output "registry_token" {
  value = selectel_craas_token_v1.token_1.token
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

* `token_ttl` - (Optional) Token lifetime. Changing this creates a new token. Available values are `1y` for a year and `12h` for 12 hours. The default value is `1y`.

## Attributes Reference

* `username` - (Sensitive) Username to access Container Registry.

* `token` - (Sensitive) Token to access Container Registry.