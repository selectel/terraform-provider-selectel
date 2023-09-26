---
layout: "selectel"
page_title: "Selectel: selectel_vpc_floatingip_v2"
sidebar_current: "docs-selectel-resource-vpc-floatingip-v2"
description: |-
  Creates and manages a public IP address for Selectel products using public API v2.
---

# selectel\_vpc\_floatingip_v2

Creates and manages a public IP address using public API v2. For more information about public IP addresses, see the [official Selectel documentation](https://docs.selectel.ru/cloud/servers/networks/about-networks/).

## Example Usage

```hcl
resource "selectel_vpc_floatingip_v2" "floatingip_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-1"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new public IP address. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Required) Pool where the public IP address is located, for example, `ru-3`. Changing this creates a new public IP address. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/).

## Attributes Reference

* `port_id` - Unique identifier of the associated OpenStack port. Learn more about the [openstack_networking_port_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_port_v2) resource in the official OpenStack documentation.

* `floating_ip_address` - Public IP address.

* `fixed_ip_address` -  Fixed private IP address of the OpenStack port, that is associated with the public IP address. Learn more about the [openstack_networking_port_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_port_v2) resource in the official OpenStack documentation.

* `status` - Status of the public IP address.

* `servers` - Cloud server that use the public IP address.

  * `id` - Unique identifier of the cloud server.

  * `name` - Name of the cloud server.

  * `status` - Status of the cloud server.

## Import

You can import a public IP address:

```shell
<<<<<<< HEAD
terraform import selectel_vpc_floatingip_v2.floatingip_1 <public_ip_id>
=======
$ export OS_DOMAIN_NAME=999999
$ export OS_USERNAME=example_user
$ export OS_PASSWORD=example_password
$ terraform import selectel_vpc_floatingip_v2.floatingip_1 <public_ip_id>
>>>>>>> upstream/master
```

where `<public_ip_id>` is a unique identifier of the public IP address, for example, `0635d78f-57a7-1a23-bf9d-9e10`. To get the public IP address ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Public IP addresses** tab. The ID is under the public IP address. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/cloud/servers/tools/openstack/) command `openstack floating ip list` and copy `ID` field.

### Environment Variables

For import, you must set the environment variable `SEL_TOKEN=<selectel_api_token>`,

where `<selectel_api_token>` is a Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).
