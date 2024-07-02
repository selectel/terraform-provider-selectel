---
layout: "selectel"
page_title: "Selectel: selectel_vpc_floatingip_v2"
sidebar_current: "docs-selectel-resource-vpc-floatingip-v2"
description: |-
  Creates and manages a public IP address for Selectel products using public API v2.
---

# selectel\_vpc\_floatingip_v2

Creates and manages a public IP address using public API v2. For more information about public IP addresses, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/servers/networks/about-networks/).

## Example Usage

```hcl
resource "selectel_vpc_floatingip_v2" "floatingip_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-1"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new public IP address. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the public IP address is located, for example, `ru-3`. Changing this creates a new public IP address. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

## Attributes Reference

* `port_id` - Unique identifier of the associated OpenStack port. Learn more about the [openstack_networking_port_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_port_v2) resource in the official OpenStack documentation.

* `floating_ip_address` - Public IP address.

* `fixed_ip_address` - Fixed private IP address of the OpenStack port, that is associated with the public IP address. Learn more about the [openstack_networking_port_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_port_v2) resource in the official OpenStack documentation.

* `status` - Status of the public IP address.

* `servers` - Cloud server that use the public IP address.

  * `id` - Unique identifier of the cloud server.

  * `name` - Name of the cloud server.

  * `status` - Status of the cloud server.

## Import

You can import a public IP address:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_vpc_floatingip_v2.floatingip_1 <public_ip_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

*  `<public_ip_id>` — Unique identifier of the public IP address, for example, `0635d78f-57a7-1a23-bf9d-9e10`. To get the public IP address ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Public IP addresses** tab. The ID is under the public IP address. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack floating ip list` and copy `ID` field.
