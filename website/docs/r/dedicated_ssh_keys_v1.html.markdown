---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_ssh_keys_v1"
sidebar_current: "docs-selectel-resource-dedicated-ssh-keys-v1"
description: |-
  Adds and manages a public SSH key for dedicated servers using public API v1.
---

# selectel_dedicated_ssh_keys_v1

Adds and manages a public SSH key for dedicated servers using public API v1. For more information how to create an SSH key pair, see the [official Selectel documentation.](https://docs.selectel.ru/en/dedicated/manage/create-and-place-ssh-key/#create-ssh-key)

## Example Usage

```hcl
resource "selectel_dedicated_ssh_keys_v1" "ssh_key_1" {
  name       = "ssh_key_1"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD..."
}
```

## Argument Reference

* `name` - (Required) Name of the SSH key. Changing this creates a new SSH key. 

* `public_key` - (Required) Public SSH key string. Changing this creates a new SSH key. Leading and trailing whitespace is ignored. 

## Attributes Reference

* `id` - Unique identifier of the SSH key. 
* `name` - Name of the SSH key.
* `public_key` - SSH public key.

## Import

You can import an SSH public key:

```bash
terraform import selectel_dedicated_ssh_keys_v1.ssh_key_1 <ssh_key_name>
```

where:

* `<ssh_key_name>` — Name of the SSH key. If multiple SSH keys with the same name exist, the first one found will be imported. We recommend to use unique names for SSH keys.
