---
layout: "selectel"
page_title: "Selectel: selectel_vpc_router_interface_v2"
sidebar_current: "docs-selectel-resource-vpc-router-interface-v2"
description: |-
  Creates and manages an interface that attaches a cloud router to a network for Selectel products using public API v2.
---

# selectel\_vpc\_router\_interface\_v2

Creates and manages an interface that attaches a cloud router to a network using public API v2. For more information about cloud routers, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/cloud-routers/).

The interface attaches the router either through a subnet or through an existing port. The router routes traffic for every subnet attached to it. The router itself is managed with the [selectel_vpc_router_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2) resource.

~> **Note:** Detaching an interface deletes its port — both a port the API created for `subnet_id` and a port you passed as `port_id`.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

### Attaching through a subnet — the API creates the port

```hcl
resource "selectel_vpc_router_interface_v2" "interface_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  router_id  = selectel_vpc_router_v2.router_1.id
  subnet_id  = selectel_vpc_network_subnet_v2.subnet_1.id
}
```

### Attaching through a port you created yourself

```hcl
resource "selectel_vpc_router_interface_v2" "interface_2" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  router_id  = selectel_vpc_router_v2.router_1.id
  port_id    = selectel_vpc_port_v2.port_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new interface. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the router is located, for example, `ru-3`. Changing this creates a new interface. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `router_id` - (Required) Unique identifier of the router the network is attached to. Changing this creates a new interface. Retrieved from the [selectel_vpc_router_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2) resource.

* `subnet_id` - (Optional) Unique identifier of the subnet to attach. You must specify either `subnet_id` or `port_id`, but not both. When you attach through `subnet_id` API creates the port for the interface. If you attach through `port_id`, `subnet_id` is computed from the fixed IP of the port and is available in the state. Changing this creates a new interface. Retrieved from the [selectel_vpc_network_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) resource.

* `port_id` - (Optional) Unique identifier of the port to attach. You must specify either `subnet_id` or `port_id`, but not both. A port passed as `port_id` must have exactly one fixed IP. If you attach through `subnet_id`, `port_id` is computed from the port the API creates and is available in the state — it matches the resource identifier used for import. Changing this creates a new interface. Retrieved from the [selectel_vpc_port_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2) resource. Destroying the interface deletes the port as well: the `selectel_vpc_port_v2` resource that created it plans to recreate the port on the next apply, and a replacement of the interface cannot reuse the deleted port.

* `force_destroy` - (Optional) Specifies whether the interface may be detached by first removing the router's static routes that point into the attached subnet. Boolean flag, the default value is `false`. Can be changed without replacement — it only steers this resource's destroy. The API refuses to detach an interface whose subnet is still in use — for example, when the router has static routes whose `next_hop` is an address inside that subnet:

  * `false` (default) — the destroy fails with the API error. Remove the routes first or set `force_destroy` to `true`.

  * `true` — the routes whose `next_hop` belongs to the subnet are removed first, and then the interface is detached.

## Attributes Reference

* `network_id` - Unique identifier of the network containing the interface port. It is computed from the actual port after create, read, or import.

## Import

You can import an interface by the identifier of its port — whether you supplied that port as `port_id` or the API created it for `subnet_id`:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_router_interface_v2.interface_1 <port_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/user-types).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/). The import ID does not contain the project, so `INFRA_PROJECT_ID` must be set.

* `<pool>` — Pool where the router is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/). The import ID does not contain the pool, so `INFRA_REGION` must be set.

* `<port_id>` — Unique identifier of the port the router is attached through, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the port ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ the network page ⟶ the **Ports** tab ⟶ copy the ID of the required port. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack port list --router <router_id>`, where `<router_id>` is the unique identifier of the router, and copy the `ID` field.

~> **Note:** A port with more than one fixed IP carries more than one router interface and cannot be imported as a single resource — the import is rejected. The import verifies that the port is attached to a device, but it does not verify that the device is a router. Use a port returned by the filtered CLI command above.
