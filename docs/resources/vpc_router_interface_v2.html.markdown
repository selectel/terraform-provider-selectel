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

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

Attaching through a subnet — the API creates the port:

```hcl
resource "selectel_vpc_router_interface_v2" "interface_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  router_id  = selectel_vpc_router_v2.router_1.id
  subnet_id  = selectel_vpc_network_subnet_v2.subnet_1.id
}
```

Attaching through a port you created yourself:

```hcl
resource "selectel_vpc_router_interface_v2" "interface_2" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  router_id  = selectel_vpc_router_v2.router_1.id
  port_id    = selectel_vpc_port_v2.port_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new interface. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the router is located, for example, `ru-3`. Changing this creates a new interface. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `router_id` - (Required) Unique identifier of the router the network is attached to. Changing this creates a new interface.

* `subnet_id` - (Optional) Unique identifier of the subnet to attach. Conflicts with `port_id`. When you attach through `port_id` instead, this is computed from the port. Changing this creates a new interface.

* `port_id` - (Optional) Unique identifier of the port to attach. Conflicts with `subnet_id`. When you attach through `subnet_id` instead, this is the port the API created for the interface. A port passed as `port_id` must have exactly one fixed IP. Changing this creates a new interface.

* `force_destroy` - (Optional) Specifies whether the interface may be detached by first removing the router's static routes that point into the attached subnet. The default value is `false`. Can be changed without replacement — it only steers this resource's destroy.

## Attributes Reference

* `network_id` - Unique identifier of the network containing the interface port. It is computed from the actual port after create, read, or import.

The identifier of the resource is the identifier of the port the router is attached through — whether you supplied that port or the API created it for the subnet.

## Import

You can import an interface by the identifier of its port:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_router_interface_v2.interface_1 <port_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). The import ID does not contain the project, so `INFRA_PROJECT_ID` must be set.

* `<pool>` — Pool where the router is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/). The import ID does not contain the pool, so `INFRA_REGION` must be set.

* `<port_id>` — Unique identifier of the port the router is attached through, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. No composite format and no `router_id` are needed: `router_id`, `network_id`, and `subnet_id` are read from the port. To get the port ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ the network page ⟶ the **Ports** tab ⟶ copy the ID of the required port. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack port list` and copy the `ID` field.

~> **Note:** If the imported port has more than one fixed IP, `subnet_id` is left unset — there is no single subnet to report. A port that is not attached to a router is rejected as not being a router interface.

## Behavior notes

Detaching an interface deletes the port the API created for it. If the interface is detached outside Terraform, or its port is deleted, the resource leaves the state on the next refresh without an error and the next plan re-creates it.

The API refuses to detach an interface whose subnet is still in use — for example, when the router has static routes whose `next_hop` is an address inside that subnet. With `force_destroy = false`, the detach is retried until the delete timeout. With `force_destroy = true`, the routes whose `next_hop` belongs to the subnet are removed first, and then the interface is detached.
