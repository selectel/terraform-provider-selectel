---
layout: "selectel"
page_title: "Selectel: selectel_vpc_keypair_v2"
sidebar_current: "docs-selectel-resource-vpc-keypair-v2"
description: |-
  Manages a V2 keypair resource within Selectel VPC.
---

# selectel\_vpc\_keypair_v2

Manages a V2 keypair resource within Selectel VPC.

## Example Usage

```hcl
resource "selectel_vpc_user_v2" "user_1" {
  password = "secret"
}

resource "selectel_vpc_keypair_v2" "keypair_tf_acc_test_1" {
  public_key = "${file("~/.ssh/id_rsa.pub")}"
  user_id    = "${selectel_vpc_user_v2.user_1.id}"
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
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selectel_vpc_keypair_v2.keypair_1 <user_id>/<name>
```
