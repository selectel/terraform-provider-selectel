---
layout: "selectel"
page_title: "Upgrading Terraform Selectel Provider to version 5.0.0"
sidebar_current: "docs-selectel-guide-iam-migrating-guide"
description: |-
 How to upgrade Terraform Selectel Provider to version 5.0.0.
---

# Upgrading Terraform Selectel Provider to version 5.0.0

In version 5.0.0, Terraform Selectel Provider introduces new resources for managing:

- service users — [selectel_iam_serviceuser_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1);
- control panel users (local and federated) — [selectel_iam_user_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_user_v1).

The selectel_iam_serviceuser_v1 resource replaces selectel_vpc_user_v2 and selectel_vpc_role_v2 resources. Update your files according to the guide.

Before upgrading to version 5.0.0, [upgrade to the most recent 4.X version of the provider](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_4) and ensure that your environment successfully runs `terraform plan`. You should not see changes you do not expect or deprecation notices.

1. In the Terraform configuration, update the version constraints:

    ```hcl
    terraform {
    required_providers {
      selectel  = {
        source  = "selectel/selectel"
        version = "~> 5.0"
      }
      openstack = {
        source  = "terraform-provider-openstack/openstack"
        version = "1.54.0"
      }
    }
    }
    ```

2. To download the new version, initialize the Terraform configuration.

    ```bash
    terraform init -upgrade
    ```

3. Backup the `.tfstate` file.
4. Remove the selectel_vpc_user_v2 resource from the `.tfstate` file:

    ```bash
    terraform state rm $(terraform state list | grep selectel_vpc_user_v2)
    ```

5. Remove the selectel_vpc_role_v2 resource from the `.tfstate` file:

    ```bash
    terraform state rm $(terraform state list | grep selectel_vpc_role_v2)
    ```

6. In the configuration files (`.tf` files), rename the selectel_vpc_user_v2 resource to selectel_iam_serviceuser_v1 and add the role of the service user to the selectel_iam_serviceuser_v1 resource. You can add multiple roles — each role in a separate block.

    ```hcl
    resource "selectel_iam_serviceuser_v1" "serviceuser_1" {
      name     = "username"
      password = "password"
      role {
        role_name = "member"
        scope     = "account"
      }
      role {
        role_name = "iam_admin"
        scope     = "account"
      }
    }
    ```

  For more information about available roles, see the [selectel_iam_serviceuser_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1) resource.

7. Remove the selectel_vpc_role_v2 resource from the configuration files.
8. [Import the service user into the `.tfstate` file](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1#import).

  After the import, the `.tfstate` file contains the data from the selectel_iam_serviceuser_v1 resource. The value in the `password` field is `UNDEFINED_WHILE_IMPORTING`. Terraform adds the password when you apply the changes.
9. To ensure that Terraform applies the required changes, preview the changes:

    ```bash
    terraform plan
    ``` 

10. If Terraform shows that Terraform will destroy a role you need, check if the `role` blocks in the selectel_iam_serviceuser_v1 resource contain all the required roles.
11. Repeat steps 6-10 for all service users.
12. If you refer to the selectel_vpc_user_v2 resource in other resources, replace it with the selectel_iam_serviceuser_v1 resource.
13. Apply the changes:

    ```bash
    terraform apply
    ```
