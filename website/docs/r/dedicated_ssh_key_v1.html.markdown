---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_ssh_keys_v1"
sidebar_current: "docs-selectel-resource-dedicated-ssh-keys-v1"
description: |-
  Creates and manages a V1 SSH key for dedicated servers that can be used with Selectel services.
---

# selectel_dedicated_ssh_keys_v1

Use this resource to create and manage SSH keys specifically intended for use with Selectel dedicated servers.

## Example Usage

```hcl
resource "selectel_iam_serviceuser_v1" "user_1" {
  name     = "tf-user"
  password = "password"

  role {
    role_name = "member"
    scope     = "account"
  }
}

resource "selectel_dedicated_ssh_keys_v1" "ssh_key_1" {
  name       = "my-dedicated-ssh-key"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD..."
  user_id    = selectel_iam_serviceuser_v1.user_1.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SSH key. Changing this creates a new SSH key.

* `public_key` - (Required) The public SSH key string. Changing this creates a new SSH key.

* `user_id` - (Required) The UUID of the user for whom the SSH key is created. Changing this creates a new SSH key.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the SSH key.
* `name` - The name of the SSH key.
* `public_key` - The public SSH key string.
* `user_id` - The UUID of the user for whom the SSH key is created.
