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

In most cases [OpenStack Terraform provider](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest)
will have to be used as well — to control resources available via native OpenStack API.

Read our [Getting Started with Terraform at Selectel](https://kb.selectel.com/docs/selectel-cloud-platform/main-services/instructions/how_to_use_terraform/)
guide to learn more.

## Example Usage

```hcl
# Configure the Selectel Provider
provider "selectel" {
  domain_name = "999999"
  username = "example_user"
  password = "example_password"
}

# Create a project
resource "selectel_vpc_project_v2" "project_1" {
  # ...
}
```

## Configuration Reference

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

* `token` - (Optional) The Selectel API key token. If omitted, the `SEL_TOKEN`
  environment variable is used. Now used only for DNS-API.

* `project_id` - (Optional) The Selectel VPC project. Used only to import
  resources that need an auth token in the project scope. If omitted,
  the `SEL_PROJECT_ID` environment variable is used.

* `region` - (Optional) The Selectel VPC region. Used only to import resources
  associated with the specific region. If omitted, the `SEL_REGION` environment
  variable is used.


## Additional Logging

To enable debug logging, set the `TF_LOG` environment variable to `DEBUG`:

```shell
$ env TF_LOG=DEBUG terraform apply
```

## Testing and Development

In order to run the Acceptance Tests for development you need to set
the `SEL_TOKEN` environment variable:

```shell
$ export OS_DOMAIN_NAME=999999
$ export OS_USERNAME=example_user
$ export OS_PASSWORD=example_password
$ env TF_ACC=1 go test -v ./selectel/...
```

Please create an issue describing a new feature or bug prior creating a pull
request.
