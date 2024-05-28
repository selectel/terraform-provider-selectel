---
layout: "selectel"
page_title: "Upgrading Terraform Selectel Provider to version 4.0.0"
sidebar_current: "docs-selectel-guide-upgrade-guide"
description: |-
  How to upgrade Terraform Selectel Provider to version 4.0.0.
---

# Upgrading Terraform Selectel Provider to version 4.0.0 

To upgrade Terraform Selectel Provider version to the new major version 4.0.0:

1. In the [Control Panel](https://my.selectel.ru/iam/users_management/users?type=service), create a service user with an Account Administrator role. Learn more [how to create a service user](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/add-user/).
2. Change the authorization block.

    From:

    ```hcl
    provider "selectel" {
      token = <token>
    }
    ```

    To: 

    ```hcl
    provider "selectel" {
      domain_name = <account_id>
      username    = <username>
      password    = <password>
    }
    ```

    where:

    * `<account_id>` - (Required) Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). For import, use the value in the `OS_DOMAIN_NAME` environment variable. Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

    * `<username>` - (Required) Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. For import, use the value in the `OS_USERNAME` environment variable. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/) and [how to create service user](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/add-user/#add-service-user).

    * `<password>` - (Required, Sensitive) Password of the service user. For import, use the value in the `OS_PASSWORD` environment variable.