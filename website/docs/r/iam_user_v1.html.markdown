---
layout: "selectel"
page_title: "Selectel: selectel_iam_euser_v1"
sidebar_current: "docs-selectel-resource-iam-user-v1"
description: |-
  Creates and manages a user for Selectel products using public API v1.
---

# selectel\_iam\_user\_v1

Creates and manages a user using public API v1. Selectel products support Identity and Access Management (IAM). For more information about users, see the [official Selectel documentation](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

## Example Usage

```hcl
resource "selectel_iam_user_v1" "user_1" {
  email       = "mail@example.com"
  role {
    role_name = "member"
    scope     = "account"
  }
}
```

## Argument Reference

* `email` - (Required) Email, to which authentication instructions will be sent. Changing this creates a new user.

* `auth_type` - (Optional) Authentication type of this user. Available values are `local` and `federated`. The default value is `local`. Changing this creates a new user.

* `federation` - (Optional) Federation info. `auth_type` must be set to `federated`.

    * `id` - (Required) Federation id.

    * `external_id` - (Required) User id on the side of Identity Provider.

* `role` - (Optional) Block, which manages roles for the service user. There can be several blocks for assigning several roles.

    * `role_name` - (Required) Role name. Available values are: `iam_admin`, `member`, `reader`, `billing`.

    * `scope` - (Required) Scope of the applied role. Available values are: `project`, and `account`.

    * `project_id` - (Optional) Project id, to which this role will be applied. `scope` must be set to `project`. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource.

## Attributes Reference

* `keystone_id` - Keystone id of the created user.



## Import

You can import a user:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_iam_user_v1.user_1 <user_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the Service User. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the Service User.

* `<user_id>` — Unique identifier (not the Keystone id!) of the user to import, for example, `123456_5432`. To get the id, use either [iam-go](https://github.com/selectel/iam-go) or [IAM API](https://developers.selectel.ru/docs/control-panel/iam/).