---
layout: "selectel"
page_title: "Selectel: selectel_iam_serviceuser_v1"
sidebar_current: "docs-selectel-resource-iam-serviceuser-v1"
description: |-
  Creates and manages a service user for Selectel products using public API v1.
---

# selectel\_iam\_serviceuser\_v1

Creates and manages a service user using public API v1. Selectel products support Identity and Access Management (IAM). For more information about service users, see the [official Selectel documentation](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

~> **Note:** The service user password is stored as raw data in a plain-text file. Learn more about [sensitive data in
state](https://developer.hashicorp.com/terraform/language/state/sensitive-data).

## Example Usage

```hcl
resource "selectel_iam_serviceuser_v1" "serviceuser_1" {
  name        = "username"
  password    = "password"
  role {
    role_name = "member"
    scope     = "account"
  }
}
```

## Argument Reference

* `name` - (Required) Name of the service user.

* `password` - (Required, Sensitive) Password of the service user.

* `role` - (Optional) Block, which manages roles for the service user. There can be several blocks for assigning several roles.

    * `role_name` - (Required) Role name. Available values are: `iam_admin`, `member`, `reader`, `billing`, `object_storage:admin`, and `object_storage_user`.

    * `scope` - (Required) Scope of the applied role. Available values are: `project`, and `account`.

    * `project_id` - (Optional) Project id, to which this role will be applied. `scope` must be set to `project`. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource.

* `enabled` - (Optional) Specifies if you can create a Cloud Platform Keystone token for the service user. Boolean flag, the default value is `true`. Learn more about [Cloud Platform Keystone tokens](https://developers.selectel.ru/docs/control-panel/authorization/#токен-для-облачной-платформы-selectel).


## Import

~> **Note:** For a guide on how to migrate from deprecated `selectel_vpc_user_v2` and `selectel_vpc_role_v2` to `selectel_iam_serviceuser_v1` follow [this link](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/migrating_to_iam_serviceuser).

You can import a service user:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_iam_serviceuser_v1.serviceuser_1 <user_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<user_id>` — Unique identifier of the service user to import, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the id, in the top right corner of the [Control panel](https://my.selectel.ru/), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the ID under the user name.