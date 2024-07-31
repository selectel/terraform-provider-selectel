---
layout: "selectel"
page_title: "Selectel: selectel_iam_group_v1"
sidebar_current: "docs-selectel-resource-iam-group-v1"
description: |-
  Creates and manages a user group for Selectel products using public API v1.
---

# selectel\_iam\_group\_v1

Creates and manages a user group for Selectel products using public API v1.
Selectel products support Identity and Access Management (IAM).
For more information about user groups, see the [official Selectel documentation](https://docs.selectel.ru/control-panel-actions/users-and-roles/groups/).

## Example Usage

```hcl
resource "selectel_iam_group_v1" "group_1" {
  name        = "My group"
  description = "My test group"
  role {
    role_name = "member"
    scope     = "account"
  }
}
```

## Argument Reference

* `name` - (Required) Group name.

* `description` - (Optional) Group description.

* `role` - (Optional) Manages group roles. You can add multiple roles – each role in a separate block. For more information about roles, see the [Roles](#roles) section.

    * `role_name` - (Required) Role name. Available role names are `iam_admin`, `member`, `reader`, and `billing`.

    * `scope` - (Required) Scope of the role. Available scopes are `account` and `project`. If `scope` is `project`, the `project_id` argument is required.

    * `project_id` - (Optional) Unique identifier of the associated project. If `scope` is `project`, the `project_id` argument is required. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

### Roles

To assign roles, use the following values for `scope` and `role_name`:

* Account administrator - `scope` is `account`, `role_name` is `member`.

* Billing administrator - `scope` is `account`, `role_name` is `billing`.

* User administrator - `scope` is `account`, `role_name` is `iam_admin`.

* Project administrator - `scope` is `project`, `role_name` is `member`.

* Account viewer - `scope` is `account`, `role_name` is `reader`.

* Project viewer - `scope` is `project`, `role_name` is `reader`.

## Import

You can import a group:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_iam_group_v1.group_1 <group_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service Users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<group_id>` — Unique identifier of the group, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the group ID, use either [iam-go](https://github.com/selectel/iam-go) or [IAM API](https://developers.selectel.ru/docs/control-panel/iam/).
