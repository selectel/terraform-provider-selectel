---
layout: "selectel"
page_title: "Selectel: selectel_iam_s3_credentials_v1"
sidebar_current: "docs-selectel-resource-iam-s3-credentials-v1"
description: |-
  Creates and manages S3 credentials for a service user using public API v1.
---

# selectel\_iam\_s3_credentials\_v1

Creates and manages S3 credentials for a service user using public API v1. S3 credentials are required to access Selectel Object Storage via S3 API. S3 credentials include Access Key and Secret Key. For more information about S3 сredentials, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/object-storage/manage/manage-access/#issue-s3-key).

~> **Note:** In S3 credentials, the Secret Key is stored as raw data in a plain-text file. Learn more about [sensitive data in state](https://developer.hashicorp.com/terraform/language/state/sensitive-data).

## Example Usage

```hcl
resource "selectel_iam_s3_credentials_v1" "s3_credentials_1" {
  user_id    = selectel_iam_serviceuser_v1.serviceuser_1.id
  project_id = selectel_vpc_project_v2.project_1.id
  name       = "S3Credentials"
}
```

## Argument Reference

* `user_id` - (Required) Unique identifier of the service user. Changing this creates new credentials. Retrieved from the [selectel_iam_serviceuser_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1) resource. Learn more about [Service Users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates new credentials. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `name` - (Required) Name of the S3 credentials. Changing this creates new credentials.

## Attributes Reference

* `access_key` - Access Key.

* `secret_key` - Secret Key.

## Import

You can import S3 credentials:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export OS_S3_CREDENTIALS_USER_ID=<user_id>
terraform import selectel_iam_s3_credentials_v1.s3_credentials_1 <access_key>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service Users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<user_id>` — Unique identifier of the service user who owns S3 credentials, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the ID, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the ID under the user name.

* `<access_key>` — Access Key from S3 сredentials. To get the Access Key, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ click on the service user who owns credentials ⟶ copy the Access Key in the **S3 keys** section.
