---
layout: "selectel"
page_title: "Selectel: selectel_iam_ec2_v1"
sidebar_current: "docs-selectel-resource-iam-ec2-v1"
description: |-
  Creates and manages a EC2 credentials for Selectel service user using public API v1.
---

# selectel\_iam\_ec2\_v1

Creates and manages a EC2 credentials for Selectel Service User using public API v1. For more information about EC2 Credentials, see the [official Selectel documentation](https://docs.selectel.ru/cloud/object-storage/manage/manage-access/#issue-s3-key).

~> **Note:** The secret key of created EC2 credential is stored as raw data in a plain-text file. Learn more about [sensitive data in state](https://developer.hashicorp.com/terraform/language/state/sensitive-data).

## Example Usage

```hcl
resource "selectel_iam_ec2_v1" "ec2_credential_1" {
  user_id     = selectel_iam_serviceuser_v1.serviceuser_1.id
  project_id  = selectel_vpc_project_v2.project_1.id
  name        = "MyCredential"
}
```

## Argument Reference

* `user_id` - (Required) A service user id to create EC2 credential for. Retrieved from the [selectel_iam_serviceuser_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1) resource. Changing this creates a new credential.

* `project_id` - (Required) A Project ID to create EC2 credential for. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Changing this creates a new credential.

* `name` - (Required) Name of the EC2 credential. Changing this creates a new credential.

## Attributes Reference

* `access_key` - Access Key ID.

* `secret_key` - Secret Access Key.