---
layout: "selectel"
page_title: "Selectel: selectel_vpc_role_v2"
sidebar_current: "docs-selectel-resource-vpc-role-v2"
description: |-
  Creates and manages a Project Administrator role for Selectel service users using public API v2.
---

# selectel\_vpc\_role_v2

Creates and manages a Project Administrator role for service users using public API v2. Selectel products support Identity and Access Management (IAM). For more information about roles, see the [official Selectel documentation](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

The role is assigned to the service user information about whom is retrieved from the [selectel_vpc_user_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_user_v2) resource.

## Example Usage

```hcl
resource "selectel_vpc_role_v2" "role__1" {
  project_id = selectel_vpc_project_v2.project_1.id
  user_id    = selectel_vpc_user_v2.user_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new role. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `user_id` - (Required) Unique identifier of the associated service user. Changing this creates a new role. Retrieved from the [selectel_vpc_user_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_user_v2) resource.

## Import

You can import a role:

```shell
<<<<<<< HEAD
terraform import selectel_vpc_role_v2.role_1 <project_id>/<user_id>
=======
$ export OS_DOMAIN_NAME=999999
$ export OS_USERNAME=example_user
$ export OS_PASSWORD=example_password
$ terraform import selectel_vpc_role_v2.role_1 <project_id>/<user_id>
>>>>>>> upstream/master
```

where:

* `<project_id>` — Unique identifier of the Cloud Platform project, for example, `a07abc12310546f1b9291ab3013a7d75`. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/), go to the **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project.

* `<user_id>` — Unique identifier of the associated service user, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the ID, in the top right corner of the [Control panel](https://my.selectel.ru/), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the ID under the user name.

### Environment Variables

For import, you must set the environment variable `SEL_TOKEN=<selectel_api_token>`,

<<<<<<< HEAD
where `<selectel_api_token>` is a Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).
=======
where `<selectel_api_token>` is a Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).
>>>>>>> upstream/master
