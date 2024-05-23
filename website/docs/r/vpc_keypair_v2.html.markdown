---
layout: "selectel"
page_title: "Selectel: selectel_vpc_keypair_v2"
sidebar_current: "docs-selectel-resource-vpc-keypair-v2"
description: |-
  Creates and manages a SSH key pair for Selectel products using public API v2.
---

# selectel\_vpc\_keypair_v2

Creates and manages a SSH key pair using public API v2. For more information about SSH key pairs, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/servers/manage/create-and-place-ssh-key/).

Selectel products support Identity and Access Management (IAM). Only service users can use SSH key pairs. To create a service user, use the [selectel_iam_serviceuser_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1) resource. For more information about service users, see the [official Selectel documentation](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

## Example Usage

```hcl
resource "selectel_vpc_keypair_v2" "keypair_1" {
  name       = "keypair"
  public_key = file("~/.ssh/id_rsa.pub")
  user_id    = selectel_iam_serviceuser_v1.user_1.id
}
```

## Argument Reference

* `name` - (Required) Name of the SSH key pair. Changing this creates a new key pair.

* `public_key` - (Required) Pregenerated OpenSSH-formatted public key. Changing this creates a new key pair. Learn more [how to create SSH key pair](https://docs.selectel.ru/en/cloud/servers/manage/create-and-place-ssh-key/#create-ssh-keys).

* `user_id` - (Required) Unique identifier of the associated service user. Changing this creates a new key pair. Retrieved from the [selectel_iam_serviceuser_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1) resource.

* `regions` - (Optional) List of pools where the key pair is located, for example, `ru-3`. Changing this creates a new key pair. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

## Import

You can import a SSH key pair:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_vpc_keypair_v2.keypair_1 <user_id>/<keypair_name>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<user_id>` — Unique identifier of the associated service user, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the ID, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user.

* `<keypair_name>` — Name of the key pair, for example, `Key`. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ the user page. The SSH key pair name is in the **SSH keys** section.