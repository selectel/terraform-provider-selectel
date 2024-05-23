---
layout: "selectel"
page_title: "Selectel: selectel_vpc_role_v2"
sidebar_current: "docs-selectel-resource-vpc-role-v2"
description: |-
  Creates and manages a Project Administrator role for Selectel service users using public API v2.
---

# selectel\_vpc\_role_v2

> **WARNING**: This resource is deprecated. Since version 5.0.0, replace the resource with the roles block in the [selectel_iam_serviceuser_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1) resource. For more information about upgrading to version 5.0.0, see the [upgrading guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_5).

Creates and manages a Project Administrator role for service users using public API v2. Selectel products support Identity and Access Management (IAM). For more information about roles, see the [official Selectel documentation](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

The role is assigned to the service user information about whom is retrieved from the [selectel_vpc_user_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_user_v2) resource.

## Example Usage

```hcl
resource "selectel_vpc_role_v2" "role__1" {
  project_id = selectel_vpc_project_v2.project_1.id
  user_id    = selectel_vpc_user_v2.user_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new role. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `user_id` - (Required) Unique identifier of the associated service user. Changing this creates a new role. Retrieved from the [selectel_vpc_user_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_user_v2) resource.

## Import

You can import a role:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_vpc_role_v2.role_1 <project_id>/<user_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the project, for example, `a07abc12310546f1b9291ab3013a7d75`. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project.

* `<user_id>` — Unique identifier of the associated service user, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the ID, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the ID under the user name.
