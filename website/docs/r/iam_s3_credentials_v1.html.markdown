---
layout: "selectel"
page_title: "Selectel: selectel_iam_s3_credentials_v1"
sidebar_current: "docs-selectel-resource-iam-s3-credentials-v1"
description: |-
  Creates and manages a S3 credentials for Selectel service user using public API v1.
---

# selectel\_iam\_s3_credentials\_v1

Creates and manages a S3 credentials for Selectel Service User using public API v1. For more information about S3 Credentials, see the [official Selectel documentation](https://docs.selectel.ru/cloud/object-storage/manage/manage-access/#issue-s3-key).

~> **Note:** The _secret key_ of created S3 credentials is stored as raw data in a plain-text file. Learn more about [sensitive data in state](https://developer.hashicorp.com/terraform/language/state/sensitive-data).

## Example Usage

```hcl
resource "selectel_iam_s3_credentials_v1" "s3_credential_1" {
  user_id     = selectel_iam_serviceuser_v1.serviceuser_1.id
  project_id  = selectel_vpc_project_v2.project_1.id
  name        = "MyCredential"
}
```

## Argument Reference

* `user_id` - (Required) A service user id to create S3 credentials for. Retrieved from the [selectel_iam_serviceuser_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1) resource. Changing this creates a new credentials.

* `project_id` - (Required) A Project ID to create S3 credentials for. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Changing this creates a new credentials.

* `name` - (Required) Name of the S3 credentials. Changing this creates a new credentials.

## Attributes Reference

~> **Note:** The _access key_ of S3 credentials is stored as _id_.

* `secret_key` - Secret Access Key.

## Import

You can import an S3 credentials:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export OS_S3_CREDENTIALS_USER_ID=<user_id>
terraform import selectel_iam_s3_credentials_v1.s3_credentials_1 <access_key>
```

where:

* `<account_id>` - Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` - Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the user. Learn more about [service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` - Password of the service user.

* `<user_id>` - Unique identifier of the service user who owns S3 credentials, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the ID, in the top right corner of the [Control panel](https://my.selectel.ru/), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the ID under the user name.

* `<access_key>` - Access Key of S3 Credentials. To get the Access Key, in the top right corner of the [Control panel](https://my.selectel.ru/), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ click on the service user who owns credentials ⟶ get the Access Key of the S3 Credentials under **S3 Credentials** section.
