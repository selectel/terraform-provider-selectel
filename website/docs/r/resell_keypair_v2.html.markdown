---
layout: "selvpc"
page_title: "SelVPC: selvpc_resell_keypair_v2"
sidebar_current: "docs-selvpc-resource-resell-keypair-v2"
description: |-
  Manages a V2 keypair resource within Resell Selectel VPC.
---

# selvpc\_resell\_keypair_v2

Manages a V2 keypair resource within Resell Selectel VPC.

## Example Usage

```hcl
resource "selvpc_resell_user_v2" "user_1" {
  password = "secret"
}

resource "selvpc_resell_keypair_v2" "keypair_tf_acc_test_1" {
  public_key = "${file("~/.ssh/id_rsa.pub")}"
  user_id    = "${selvpc_resell_user_v2.user_1.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the keypair. Changing this creates a new keypair.

* `public_key` - (Required) A pregenerated OpenSSH-formatted public key.
  Changing this creates a new keypair.

* `regions` - (Optional) List of region names where keypair is need to be
  created. Keypair will be created in all available regions if omitted. Changing
  this creates a new keypair.

* `user_id` - (Required) An associated Selectel VPC user. Changing this
  creates a new keypair.

## Attributes Reference

There are no additional attributes for this resource.

## Import

Keypairs can be imported by specifying `user_id` and `name` arguments, separated
by a forward slash:

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selvpc_resell_keypair_v2.keypair_1 <user_id>/<name>
```
