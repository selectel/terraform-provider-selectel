---
layout: "selectel"
page_title: "Provider: Selectel"
sidebar_current: "docs-selectel-index"
description: |-
  Use the Selectel provider to manage Selectel products.
---

# Selectel provider

Use the Selectel provider to manage [Selectel products](https://docs.selectel.ru/).

To manage resources available via OpenStack API, use [OpenStack Terraform provider](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest).

## Example Usage

```hcl
terraform {
  required_providers {
    selectel = {
      source = "selectel/selectel"
      version = "~> 4.0.1"
    }
  }
}

# Create a Cloud Platform project
resource "selectel_vpc_project_v2" "project_1" {
  ...
}
```

## Authentication (4.0.0 and later)

```hcl
# Configure the Selectel provider

provider "selectel" {
  domain_name = "123456"
  username    = "user"
  password    = "password"
}
```

## Argument Reference (4.0.0 and later)

* `domain_name` - (Required) Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). For import, use the value in the `OS_DOMAIN_NAME` environment variable. Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `username` - (Required) Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. For import, use the value in the `OS_USERNAME` environment variable. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/) and [how to create service user](https://docs.selectel.ru/control-panel-actions/users-and-roles/add-user/#добавить-сервисного-пользователя).

* `password` - (Required, Sensitive) Password of the service user. For import, use the value in the `OS_PASSWORD` environment variable.

* `user_domain_name` - (Optional) Selectel account ID. Use only for users that were created and assigned a role in a different account. Applicable only to public cloud. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). For import, use the value in the `OS_USER_DOMAIN_NAME` environment variable.

* `auth_url`- (Optional) Keystone Identity authentication URL for authentication via user credentials. If skipped, the provider uses the default endpoint `https://cloud.api.selcloud.ru/identity/v3/`. For import, use the value in the OS_AUTH_URL environment variable.

* `auth_region` - (Optional) Pool where the endpoint for Keystone API and Resell API is located, for example, ru-3. If skipped, the provider uses the default pool `ru-1`. Does not affect the region parameter in the resources, but it is preferable to use one pool in a manifest. For import, use the value in the OS_REGION_NAME environment variable. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/).

* `project_id` - (Optional) Unique identifier of the Cloud Platform project. Use only to import resources that are associated with the specific project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/), go to the **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. As an alternative, you can retrieve project ID from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. If skipped, use the `SEL_PROJECT_ID` environment variable. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Optional) Pool, for example, `ru-3`. Use only to import resources from the specific pool. If skipped, use the `SEL_REGION` environment variable. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/).

## Authentication (up to 3.11.0)

```hcl
# Configure the Selectel provider

provider "selectel" {
  token = "Kfpfdf7fjdv0_123456"
}
```

## Argument Reference (up to 3.11.0)

* `token` - (Required) Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).

* `endpoint` - (Optional) Selectel API endpoint. Use only for test environments. If skipped, the provider automatically uses the official Selectel endpoint.

* `project_id` - (Optional) Unique identifier of the Cloud Platform project. Use only to import resources that are associated with the specific project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/), go to the **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. As an alternative, you can retrieve project ID from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/). If skipped, use the `SEL_PROJECT_ID` environment variable. 

* `region` - (Optional) Pool, for example, `ru-3`. Use only to import resources from the specific pool. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/). If skipped, use the `SEL_REGION` environment variable.
