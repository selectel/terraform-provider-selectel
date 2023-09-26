# Basic Selectel VPC project with quotas and floating IPs in different regions

This example provides a simple project with quotas in different regions.

After you run `terraform apply` on this configuration, it will create a single
project and output allocated floating ips in selected regions.

Created project and floating IPs then can be used to create OpenStack instances.
You can use [terraform-provider-openstack](https://github.com/terraform-providers/terraform-provider-openstack)
to manage OpenStack instances inside the created project.

To run this example you need to set `OS_DOMAIN_NAME` (your account id), `OS_USERNAME`, `OS_PASSWORD` variables with
authentication info that you can get from the [service users](https://my.selectel.ru/profile/users_management/users?type=service) page.

You can find additional examples in the [selectel/terraform-examples](https://github.com/selectel/terraform-examples).
