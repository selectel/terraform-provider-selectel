---
layout: "selectel"
page_title: "Selectel: selectel_iam_user_v1"
sidebar_current: "docs-selectel-resource-iam-user-v1"
description: |-
  Creates and manages a control panel user or a federated user for Selectel products using public API v1.
---

# selectel\_iam\_user\_v1

Creates and manages a control panel (local) user or a federated user using public API v1. Selectel products support Identity and Access Management (IAM). For more information about users, see the [official Selectel documentation](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

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

* `email` - (Required) Email address of the user. Changing this creates a new user. We will send authentication instructions to this email.

* `auth_type` - (Optional) Authentication type of the user. Changing this creates a new user. Available types are `local` (for control panel users, to store the credentials locally in Selectel) and `federated` (for federated users, to store the credentials in the corporate Identity Provider). The default value is `local`. If `auth_type` is `federated`, the `federation` argument is required.

* `federation` - (Optional) Information about the federation. `auth_type` must be set to `federated`.

    * `id` - (Required) Unique identifier of the federation.

    * `external_id` - (Required) Unique identifier of the user assigned by the Identity Provider.

* `role` - (Optional) Manages service user roles. You can add multiple roles – each role in a separate block. For more information about roles, see the [Roles](#roles) section.

    * `role_name` - (Required) Role name. Available role names are `iam_admin`, `member`, `reader`, and `billing`.

    * `scope` - (Required) Scope of the role. Available scopes are `account` and `project`. If `scope` is `project`, the `project_id` argument is required.

    * `project_id` - (Optional) Unique identifier of the associated project. Changing this creates a new service user. If `scope` is `project`, the `project_id` argument is required. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

### Roles

To assign roles, use the following values for `scope` and `role_name`:

* Account administrator - `scope` is `account`, `role_name` is `member`.

* Billing administrator - `scope` is `account`, `role_name` is `billing`.

* User administrator - `scope` is `account`, `role_name` is `iam_admin`.

* Project administrator - `scope` is `project`, `role_name` is `member`.

* Account viewer - `scope` is `account`, `role_name` is `reader`.

* Project viewer - `scope` is `project`, `role_name` is `reader`.

* Object storage admin - `scope` is `project`, `role_name` is `object_storage:admin`.

* Object storage user - `scope` is `project`, `role_name` is `object_storage_user`.

## Attributes Reference

* `keystone_id` - Unique Keystone identifier of the user.

## Import

You can import a user:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_iam_user_v1.user_1 <user_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service Users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<user_id>` — Unique identifier of the user to import (not the Keystone ID), for example, `123456_5432`. To get the ID, use either [iam-go](https://github.com/selectel/iam-go) or [IAM API](https://developers.selectel.ru/docs/control-panel/iam/).
