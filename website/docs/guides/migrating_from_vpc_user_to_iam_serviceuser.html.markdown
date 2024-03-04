---
layout: "selectel"
page_title: "Migrating from vpc_user_v2/vpc_roles_v2 to iam_serviceuser_v1"
sidebar_current: "docs-selectel-guide-iam-migrating-guide"
description: |-
  How to migrate from vpc_user_v2/vpc_roles_v2 to iam_serviceuser_v1.
---

# Migrating from vpc_user_v2/vpc_roles_v2 to iam_serviceuser_v1

Let’s take this example as the configuration to be migrated:

```hcl
resource "selectel_vpc_user_v2" "service_user" {
  name     = "MyServiceUser"
  password = "Qazwsxedc123!"
  enabled  = true
}

resource "selectel_vpc_role_v2" "role" {
  project_id = "1a2b3c4d..."
  user_id    = selectel_vpc_user_v2.service_user.id
}
```


To migrate from a deprecated resources _selectel_vpc_user_v2_ and _selectel_vpc_roles_v2_ to a new _selectel_iam_serviceuser_v1_ follow these steps:

1. Remove the existing resources from the _.tfstate_ file
    
    Run the following command to remove information about _selectel_vpc_user_v2_ from _.tfstate_:

    ```bash
    terraform state rm selectel_vpc_user_v2.service_user
 
    Removed selectel_vpc_user_v2.service_user
    Successfully removed 1 resource instance(s).
    ```
    Also remove information about _selectel_vpc_role_v2_ from _.tfstate_:

    ```bash
    terraform state rm selectel_vpc_role_v2.role
 
    Removed selectel_vpc_role_v2.role
    Successfully removed 1 resource instance(s).
    ```

    All necessary roles that the service user has will be later put in .tfstate during the import.

2. Update the configuration files (_.tf_ files)
    
    Change the name of your resource from _selectel_vpc_user_v2_ to _selectel_iam_serviceuser_v1_. 

    At the same time, you need to add manualy the roles that your service user has to the _selectel_iam_serviceuser_v1_.
    
    
    The _selectel_vpc_role_v2_ resource can be removed.

    For example, if the _selectel_vpc_user_v2_ has only _Project Administrator_ role (i. e. _selectel_vpc_role_v2_), then the resulting resource should look like this:

    ```hcl
    resource "selectel_iam_serviceuser_v1" "service_user" {
        name     = "username"
        password = "Qazwsxedc123!"
        enabled  = true
        role {
            role_name  = "member"
            scope      = "project"
            project_id = "1a2b3c4d..."
        }
    }
    ```

    You can add multiple roles. For example, if the _selectel_vpc_user_v2_ is, let's say, _Project Administrator_ and _IAM Administrator_, then the resulting resource should look like this:

    ```hcl
    resource "selectel_iam_serviceuser_v1" "service_user" {
        name     = "username"
        password = "Qazwsxedc123!"
        enabled  = true
        role {
            role_name  = "member"
            scope      = "project"
            project_id = "1a2b3c4d..."
        }
        role {
            role_name  = "iam_admin"
            scope      = "project"
        }
    }
    ```
3. Import the service user into a _.tfstate_ file
    
    To import an existing service user, we need to know its ID. It can be previously obtained from _.tfstate_ before running _terraform state rm_ command, or from _my.selectel_ panel, or in some other way.

    ```bash
    terraform import selectel_iam_serviceuser_v1.service_user <YOUR_SERVICE_USER_ID>
    ```

    The output should be similar to the following:
    
    ```bash
    selectel_iam_serviceuser_v1.service_user: Importing from ID "<YOUR_SERVICE_USER_ID>"...
    selectel_iam_serviceuser_v1.service_user: Import prepared!
    Prepared selectel_iam_serviceuser_v1 for import selectel_iam_serviceuser_v1.service_user: Refreshing state... [id=<YOUR_SERVICE_USER_ID>]
 
    Import successful!
 
    The resources that were imported are shown above. These resources are now in your Terraform state and will henceforth be managed by Terraform.
    ```

    After this, your _.tfstate_ will contain the _selectel_iam_serviceuser_v1_ resource information, but the **password field will be set to "IMPORT_FAILED"**, because Terraform can't retrieve it from _my.selectel_. To fix this, just call `terraform apply`:

    ```bash
    terraform apply

    ---

    Terraform will perform the following actions:

    # selectel_iam_serviceuser_v1.user will be updated in-place
    ~ resource "selectel_iam_serviceuser_v1" "service_user" {
        id       = "<YOUR_SERVICE_USER_ID>"
        name     = "MyServiceUser"
      ~ password = (sensitive value)
        # (1 unchanged attribute hidden)

        # (1 unchanged block hidden)
    }

    Plan: 0 to add, 1 to change, 0 to destroy.
    ```

    After this, the correct password will be set in _.tfstate_ for this service user and the migration from _selectel_vpc_user_v2_ and _selectel_vpc_role_v2_ to _selectel_iam_serviceuser_v1_ can be considered complete.

