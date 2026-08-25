---
layout: "selectel"
page_title: "Selectel: selectel_vpc_floatingip_associate_v2"
sidebar_current: "docs-selectel-resource-vpc-floatingip-associate-v2"
description: |-
  Creates and manages the association between a public floating IP address and a port for Selectel products using public API v2.
---

# selectel\_vpc\_floatingip\_associate\_v2

Creates and manages the association between an existing public floating IP address and an existing port using public API v2. For more information about public floating IP addresses, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/public-floating-ip-addresses/).

The resource manages only the association and never creates or deletes the public floating IP address or the port. The address is managed with the [selectel_vpc_floatingip_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_floatingip_v2) resource, and the port with the [selectel_vpc_port_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2) resource.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_floatingip_associate_v2" "association_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-6"
  floating_ip = selectel_vpc_floatingip_v2.floatingip_1.floating_ip_address
  port_id     = selectel_vpc_port_v2.port_1.id
  fixed_ip    = "192.168.199.50"
  depends_on  = [selectel_vpc_router_interface_v2.interface_1]
}
```

~> **Note:** The subnet of the port must already have a route to the external network the public floating IP address comes from. No argument of this resource refers to the router interface that provides the route, so declare the dependency with `depends_on`.

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new association. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the public floating IP address and port are located, for example, `ru-6`. Changing this creates a new association. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `floating_ip` - (Required) Public floating IP address to associate. Changing this creates a new association.

* `port_id` - (Required) Unique identifier of the existing port. Changing this updates the association in place.

* `fixed_ip` - (Optional) Fixed IP address of `port_id` to associate. If omitted, the remote API selects an address and the provider stores it in state. Changing it updates the association in place.

## Attributes Reference

No attributes are exported beyond the arguments. The resource identifier is the remote identifier of the public floating IP address, not the address itself.

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

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). The import ID does not contain the project, so `INFRA_PROJECT_ID` must be set.

* `<pool>` — Pool where the public floating IP address and port are located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/). The import ID does not contain the pool, so `INFRA_REGION` must be set.

* `<floatingip_id>` — Unique identifier of the public floating IP address, not its address, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the public floating IP address ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Public Floating IP addresses** tab ⟶ copy the ID of the required address. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack floating ip list` and copy the `ID` field.

The import fills `floating_ip`, `port_id`, and `fixed_ip` from the remote public floating IP address.

## Behavior notes

Changes made outside Terraform are reflected during refresh. If the association is detached or moved, the next plan restores the configured port and fixed IP. If the public floating IP address itself is deleted, the association leaves state.

Deleting this resource detaches the public floating IP address without deleting either the address or the port. Detaching an already detached address is idempotent. If the address itself no longer exists when the resource is deleted, the delete succeeds without an error — there is nothing left to detach.
