---
layout: "selectel"
page_title: "Provider: Selectel"
sidebar_current: "docs-selectel-index"
description: |-
  The Selectel provider is used to interact with the Selectel resources. The provider requires service user.
---

# Selectel provider

The Selectel provider is used to interact with the Selectel resources. The provider
requires service user that can be created on [users' management](https://my.selectel.ru/profile/users_management/users) page.

To interact with resources that are available via OpenStack API, you can also use [OpenStack Terraform provider](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest).

## Example Usage

```hcl
terraform {
  required_providers {
    selectel = {
      source = "selectel/selectel"
      version = "~> 3.11.0"
    }
  }

# Configure the Selectel Provider
provider "selectel" {
  domain_name = "999999"
  username = "example_user"
  password = "example_password"
}

# Create a Cloud Platform project
resource "selectel_vpc_project_v2" "project_1" {
  ...
}
```

## Authentication

The following arguments are supported:

* `username` - (Required) Service user username. Reference to OpenStack-like `OS_USERNAME` environment variable.
* `password` - (Required) Service user password. Reference to OpenStack-like `OS_PASSWORD` environment variable.
* `domain_name` - (Required) Your domain name i.e. your account id. Reference to OpenStack-like `OS_DOMAIN_NAME` environment variable.

* `user_domain_name` - (Optional) A specific field for users who were created in a different domain 
but assigned a role in a different domain. You probably don't need to use this field in public cloud.
Reference to OpenStack-like `OS_USER_DOMAIN_NAME` environment variable.

* `auth_url` - (Optional) Keystone address to authenticate via user credentials.
If omitted, the provider will use default endpoint automatically.
Reference to OpenStack-like `OS_AUTH_URL` environment variable.

* `project_id` - (Optional) The Selectel VPC project. Used only to import
  resources that need an auth token in the project scope. If omitted,
  the `SEL_PROJECT_ID` environment variable is used.

* `region` - (Optional) The Selectel VPC region. Used only to import resources
  associated with the specific region. If omitted, the `SEL_REGION` environment
  variable is used.


## Additional Logging
To enable debug logging, set the `TF_LOG` environment variable to `DEBUG`:

provider "selectel" {
  token = "<selectel_token>"
}


## Argument Reference

* `token` - (Required) Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel). If skipped, use the `SEL_TOKEN` environment variable.

* `endpoint` - (Optional) Selectel API endpoint. Use only for test environments. If skipped, the provider automatically uses the official Selectel endpoint.

* `project_id` - (Optional) Unique identifier of the Cloud Platform project. Use only to import resources that are associated with the specific project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/), go to the **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. As an alternative, you can retrieve project ID from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/). If skipped, use the `SEL_PROJECT_ID` environment variable. 

* `region` - (Optional) Pool, for example, `ru-3`. Use only to import resources from the specific pool. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/). If skipped, use the `SEL_REGION` environment variable.

In order to run the Acceptance Tests for development you need to set
auth credentials environment variables:

```shell
$ export OS_DOMAIN_NAME=999999
$ export OS_USERNAME=example_user
$ export OS_PASSWORD=example_password
$ env TF_ACC=1 go test -v ./selectel/...
```

