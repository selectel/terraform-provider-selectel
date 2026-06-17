---
layout: "selectel"
page_title: "Selectel: selectel_vpc_public_port_v1"
sidebar_current: "docs-selectel-resource-vpc-public-port-v1"
description: |-
  Creates and manages a direct public IP address (public port) for Selectel products using public API v1
---

# selectel\_vpc\_public_port\_v1

Creates and manages a direct public IP address (public port) in VPC using public API v1. For more information about direct public IP address, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/direct-public-ip-addresses).

## Example usage

```hcl
resource "selectel_vpc_public_port_v1" "port_1" {
  project_id         = selectel_vpc_project_v2.project_1.id
  region             = "ru-3"
  description        = "my-own-direct-public-ip-port"
  admin_state_up     = false
  security_group_ids = [openstack_networking_secgroup_v2.sg_1.id]
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier (no-hyphens) of the associated project. Changing this creates a new public port. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the public port is located, for example, `ru-3`. Changing this creates a new public port. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `description` - (Optional) Public port description. The default value is empty string.

* `admin_state_up` - (Optional) Enables (`true`) or disables (`false`) the public port administratively. The default value is (`true`).

* `security_group_ids` - (Optional) List of OpenStack security group identifiers to associate with the public port. Learn more about the [openstack_networking_secgroup_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_secgroup_v2) resource in the official OpenStack documentation. The default value is identifier of default security group in the project.

## Attributes Reference

* `network_id` - Identifier of the network where the public port is allocated

* `ip_address` - Direct public IP address assigned to the public port.

* `subnet` - CIDR of the subnet where the public port is allocated.

* `gateway` - IP address of the subnet gateway.

## Import

You can import a public port:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export SEL_PROJECT_ID=<selectel_project_id>
export SEL_REGION=<selectel_pool>

terraform import selectel_vpc_public_port_v1.port <public_port_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).
* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).
* `<password>` — Password of the service user.
* `<selectel_project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).
* `<selectel_pool>` — Pool where the public port is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Direct Public IP** tab ⟶ copy the name of the required pool in the dropdown list.
* `<public_port_id>` — Unique identifier of the public port, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the public port ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Direct Public IP** tab ⟶ copy the ID of the required public port.
