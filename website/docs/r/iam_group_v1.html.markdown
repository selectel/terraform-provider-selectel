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
For more information about user groups, see the [official Selectel documentation](https://docs.selectel.ru/en/access-control/groups/about-groups/).

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

* `role` - (Optional) Manages group roles. You can add multiple roles – each role in a separate block.

    * `role_name` - (Required) Role name.

    * `scope` - (Required) Scope of the role. Available scopes are `account` and `project`. If `scope` is `project`, the `project_id` argument is required.

    * `project_id` - (Optional) Unique identifier of the associated project. If `scope` is `project`, the `project_id` argument is required. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

## Import

You can import a group:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_iam_group_v1.group_1 <group_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/service-users), go to **Account** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service Users](https://docs.selectel.ru/en/access-control/user-types/).

* `<password>` — Password of the service user.

* `<group_id>` — Unique identifier of the group, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the group ID, use either [Control panel](https://my.selectel.ru/iam/groups) or [IAM API](https://docs.selectel.ru/en/api/users-and-roles/).
