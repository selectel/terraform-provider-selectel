---
layout: "selectel"
page_title: "Upgrading Terraform Selectel Provider to version 5.0.0"
sidebar_current: "docs-selectel-guide-iam-migrating-guide"
description: |-
  How to migrate from vpc_user_v2/vpc_roles_v2 to iam_serviceuser_v1.
---

# Upgrading Terraform Selectel Provider to version 5.0.0

Terraform Selectel Provider 5.0.0 introduces a new approach for working with panel users, service users and S3)-credentials through the IAM API. 

This guide can help you to migrate your current service users configurations (made with _selectel_vpc_user_v2_ and _selectel_vpc_role_v2_) to a new resources (_selectel_iam_serviceuser_v1_).

~> **Note:** Make sure you have a backup of your current _.tfstate_ file before going through the steps below.

Let’s take this example as a configuration to be migrated:

```hcl
resource "selectel_vpc_user_v2" "my_service_user" {
  name     = "MyServiceUser"
  password = "Qazwsxedc123!"
  enabled  = true
}

resource "selectel_vpc_role_v2" "role" {
  project_id = "1a2b3c4d..."
  user_id    = selectel_vpc_user_v2.my_service_user.id
}
```

To migrate from a deprecated resources _selectel_vpc_user_v2_ and _selectel_vpc_roles_v2_ to a new _selectel_iam_serviceuser_v1_ follow these steps:

1. Obtain your _my_service_user_ id. 
    
    We'll need to provide our service user id further so Terraform knows which service user should be migrated. This id can be obtained from:
        
    * _.tfstate_: find _selectel_vpc_user_v2_._my_service_user_ in your _.tfstate_ and get the _id_ field value;
        
    * _my.selectel_ panel: go to the account menu ⟶     **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the id.

2. Remove the existing resources from the _.tfstate_ file
    
    Run the following command to remove information about _selectel_vpc_user_v2_ from _.tfstate_:

    ```bash
    terraform state rm selectel_vpc_user_v2.my_service_user
 
    Removed selectel_vpc_user_v2.my_service_user
    Successfully removed 1 resource instance(s).
    ```
    Also remove information about _selectel_vpc_role_v2_ from _.tfstate_:

    ```bash
    terraform state rm selectel_vpc_role_v2.role
 
    Removed selectel_vpc_role_v2.role
    Successfully removed 1 resource instance(s).
    ```

    All necessary roles that the service user has will be later put in _.tfstate_ during the import.

3. Update the configuration files (_.tf_ files)
    
    Change the name of your resource from _selectel_vpc_user_v2_ to _selectel_iam_serviceuser_v1_. 

    At the same time, you need to add manualy the roles that your service user has to the _selectel_iam_serviceuser_v1_.
    
    The _selectel_vpc_role_v2_ resource can be removed.

    For example, if the _selectel_vpc_user_v2_ has only _Project Administrator_ role (i. e. _selectel_vpc_role_v2_), then the resulting resource should look like this:

    ```hcl
    resource "selectel_iam_serviceuser_v1" "my_service_user" {
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

    You can add multiple roles. For example, if the _selectel_vpc_user_v2_ is _Project Administrator_ in two projects, then the resulting resource should look like this:

    ```hcl
    resource "selectel_iam_serviceuser_v1" "my_service_user" {
        name     = "username"
        password = "Qazwsxedc123!"
        enabled  = true
        role {
            role_name  = "member"
            scope      = "project"
            project_id = "1a2b3c4d..."
        }
        role {
            role_name  = "member"
            scope      = "project"
            project_id = "5e6f7g8h..."
        }
    }
    ```
4. Import the service user into a _.tfstate_ file
    
    Now, provide the service user id you retrieved on the step 1: 

    ```bash
    terraform import selectel_iam_serviceuser_v1.my_service_user <YOUR_SERVICE_USER_ID>
    ```

    The output should be similar to the following:
    
    ```bash
    selectel_iam_serviceuser_v1.my_service_user: Importing from ID "<YOUR_SERVICE_USER_ID>"...
    selectel_iam_serviceuser_v1.my_service_user: Import prepared!
    Prepared selectel_iam_serviceuser_v1 for import selectel_iam_serviceuser_v1.my_service_user: Refreshing state... [id=<YOUR_SERVICE_USER_ID>]
 
    Import successful!
 
    The resources that were imported are shown above. These resources are now in your Terraform state and will henceforth be managed by Terraform.
    ```

    After this, your _.tfstate_ will contain the _selectel_iam_serviceuser_v1_ resource information, but the **password field will be set to "UNDEFINED_WHILE_IMPORTING"**, because password is not stored on the server, so Terraform can't retrieve it from _my.selectel_. To fix this, just call `terraform apply`:

    ```bash
    terraform apply

    ---

    Terraform will perform the following actions:

    # selectel_iam_serviceuser_v1.user will be updated in-place
    ~ resource "selectel_iam_serviceuser_v1" "my_service_user" {
        id       = "<YOUR_SERVICE_USER_ID>"
        name     = "MyServiceUser"
      ~ password = (sensitive value)
        # (1 unchanged attribute hidden)

        # (1 unchanged block hidden)
    }

    Plan: 0 to add, 1 to change, 0 to destroy.
    ```

    ~> **Note:** Make sure your output shows only one change (_password_). If there are any roles to be destroyed, it means, that you didn't add all necessary _role_ blocks for the roles stored in _.tfstate_ file.

    After this, the correct password will be set in _.tfstate_ for this service user and the migration from _selectel_vpc_user_v2_ and _selectel_vpc_role_v2_ to _selectel_iam_serviceuser_v1_ can be considered complete.

    Also, in case your _selectel_vpc_user_v2_ was used in other resources, such as _selectel_vpc_keypair_v2_ for example, don't forget to change it's name to _selectel_iam_serviceuser_v1_.

