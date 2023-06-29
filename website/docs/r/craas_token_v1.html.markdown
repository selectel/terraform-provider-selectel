---
layout: "selectel"
page_title: "Selectel: selectel_craas_token_v1"
sidebar_current: "docs-selectel-resource-craas-token-v1"
description: |-
Manages a V1 token resource within Selectel Container Registry Service.
---

# selectel\_craas\_token\_v1

Manages a V1 token resource within Selectel Container Registry Service.

## Basic usage example

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  name = "my-first-project"
}

resource "selectel_craas_token_v1" "token_1" {
  project_id = selectel_vpc_project_v2.project_1.id
}
```

## Docker CLI login example

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  name = "my-first-project"
}

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

The following arguments are supported:

* `project_id` - (Required) An associated Selectel VPC project.
  Changing this creates a new token.

* `token_ttl` - (Optional) Represents token expiration duration.
  Accepts "1y" or "12h". Default is "1y".
  Changing this creates a new token.

## Attributes Reference

The following attributes are exported:

* `username` - Contains a username to access container registry.
  Sensitive value.

* `token` - Contains a token to access container registry.
  Sensitive value.

## Import

Token resource import is not supported.
