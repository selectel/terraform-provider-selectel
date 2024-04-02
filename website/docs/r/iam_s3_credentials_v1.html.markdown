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