# Basic SelVPC project with quotas and floating IPs in different regions

This example provides a simple project with quotas in different regions.

After you run `terraform apply` on this configuration, it will create a single
project and output allocated floating ips in selected regions.

Created project and floating IPs then can be used to create OpenStack instances.
You can use [terraform-provider-openstack](https://github.com/terraform-providers/terraform-provider-openstack) to manage
OpenStack instances inside the created project.

To run this example you need to set `SEL_TOKEN` variable with a token key string
that you can get from the [apikeys](https://my.selectel.ru/profile/apikeys) page.
