---
layout: "selectel"
page_title: "Selectel: selectel_vpc_floatingip_associate_v2"
sidebar_current: "docs-selectel-resource-vpc-floatingip-associate-v2"
description: |-
  Creates and manages the association between a public floating IP address and a port for Selectel products using public API v2.
---

# selectel\_vpc\_floatingip\_associate\_v2

Creates and manages the association between an existing public floating IP address and an existing port using public API v2. For more information about public floating IP addresses, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/public-floating-ip-addresses/).

The resource manages only the association and never creates or deletes the public floating IP address or the port: destroying it only detaches the address from the port, and both stay. The address is managed with the [selectel_vpc_floatingip_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_floatingip_v2) resource, and the port with the [selectel_vpc_port_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2) resource.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_floatingip_associate_v2" "association_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-6"
  floating_ip = selectel_vpc_floatingip_v2.floatingip_1.floating_ip_address
  port_id     = selectel_vpc_port_v2.port_1.id
  depends_on  = [selectel_vpc_router_interface_v2.interface_1]
}
```

~> **Note:** The subnet of the port must already have a route to the external network the public floating IP address comes from. No argument of this resource refers to the router interface that provides the route, so declare the dependency with `depends_on`.

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new association. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the public floating IP address and port are located, for example, `ru-6`. Changing this creates a new association. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `floating_ip` - (Required) Public floating IP address to associate. Changing this creates a new association. Retrieved from the [selectel_vpc_floatingip_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_floatingip_v2) resource.

* `port_id` - (Required) Unique identifier of the existing port. Changing this updates the association in place. Retrieved from the [selectel_vpc_port_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2) resource.

* `fixed_ip` - (Optional) Fixed IP address of port_id to associate. Must be set if the port has multiple fixed IP addresses. If the port has a single fixed IP address, can be omitted — the provider stores the address selected by the API in state. Changing it updates the association in place.

* `timeouts` — (Optional) Timeout values. After the association is created or updated, the provider waits for the public floating IP address to become `ACTIVE`. If the address is still `DOWN` when the timeout expires, the apply finishes with a warning.
The default values are the following:
  * create = "5m",
  * update = "5m".

## Import

You can import an association by the remote identifier of its public floating IP address:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_floatingip_associate_v2.association_1 <floatingip_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/user-types).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `<pool>` — Pool where the public floating IP address and port are located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `<floatingip_id>` — Unique identifier of the public floating IP address, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the public floating IP address ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Public Floating IP addresses** tab ⟶ copy the ID of the required address. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack floating ip list` and copy the `ID` field.

The import fills `floating_ip`, `port_id`, and `fixed_ip` from the remote public floating IP address.
