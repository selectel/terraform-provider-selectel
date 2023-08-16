---
layout: "selectel"
page_title: "Selectel: selectel_vpc_keypair_v2"
sidebar_current: "docs-selectel-resource-vpc-keypair-v2"
description: |-
  Manages a V2 keypair resource within Selectel VPC.
---

# selectel\_vpc\_keypair_v2

Creates and manages a SSH key pair using public API v2. For more information about SSH key pairs, see the [official Selectel documentation](https://docs.selectel.ru/cloud/servers/manage/create-and-place-ssh-key/).

Selectel products support Identity and Access Management (IAM). Only service users can use SSH key pairs. To create a service user, use the [selectel_vpc_user_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_user_v2) resource. For more information about service users, see the [official Selectel documentation](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

## Example Usage

```hcl
resource "selectel_vpc_keypair_v2" "keypair_1" {
  name       = "keypair"
  public_key = file("~/.ssh/id_rsa.pub")
  user_id    = selectel_vpc_user_v2.user_1.id
}
```

## Argument Reference

* `name` - (Required) Name of the SSH key pair. Changing this creates a new key pair.

* `public_key` - (Required) Pregenerated OpenSSH-formatted public key. Changing this creates a new key pair. Learn more [how to create SSH key pair](https://docs.selectel.ru/cloud/servers/manage/create-and-place-ssh-key/#создать-ssh-ключи).

* `user_id` - (Required) Unique identifier of the associated service user. Changing this creates a new key pair. Retrieved from the [selectel_vpc_user_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_user_v2) resource.

* `regions` - (Optional) List of pools where the key pair is located, for example, `ru-3`. Changing this creates a new key pair. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/).

## Import

You can import a SSH key pair:

```shell
terraform import selectel_vpc_keypair_v2.keypair_1 <user_id>/<keypair_name>
```

where:

* `<user_id>` — Unique identifier of the associated service user, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the ID, in the top right corner of the [Control panel](https://my.selectel.ru/), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the ID of the required user.

* `<keypair_name>` — Name of the key pair, for example, `Key`. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ the user page. The SSH key pair name is in the **SSH keys** section.

### Environment Variables

For import, you must set the environment variable `SEL_TOKEN=<selectel_api_token>`,

where `<selectel_api_token>` is a Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).
