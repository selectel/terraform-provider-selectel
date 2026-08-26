---
layout: "selectel"
page_title: "Selectel: selectel_vpc_router_route_v2"
sidebar_current: "docs-selectel-resource-vpc-router-route-v2"
description: |-
  Creates and manages a single static route of a cloud router for Selectel products using public API v2.
---

# selectel\_vpc\_router\_route\_v2

Creates and manages a single static route of a cloud router using public API v2. For more information about static routes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/static-routes/).

The resource manages one pair of `destination_cidr` and `next_hop` and leaves every other route of the router alone. Use [selectel_vpc_router_routes_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_routes_v2) instead when you want Terraform to own the router's complete route list.

Do not manage the same router with both resources: selectel_vpc_router_routes_v2 owns the router's complete route list and removes routes added by this resource on its next apply.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_router_route_v2" "route_1" {
  project_id       = selectel_vpc_project_v2.project_1.id
  region           = "ru-6"
  router_id        = selectel_vpc_router_v2.router_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.10"
  depends_on       = [selectel_vpc_router_interface_v2.interface_1]
}
```

~> **Note:** A next hop is only accepted once the router has an interface in the subnet the address belongs to, so depend on the `selectel_vpc_router_interface_v2` resource that attaches it.

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new route. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the router is located, for example, `ru-3`. Changing this creates a new route. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `router_id` - (Required) Unique identifier of the router the route belongs to. Changing this creates a new route. Retrieved from the [selectel_vpc_router_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2) resource.

* `destination_cidr` - (Required) Destination of the route, for example, `10.0.0.0/8`. Changing this creates a new route.

* `next_hop` - (Required) Address traffic for `destination_cidr` is forwarded to. Changing this creates a new route.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is `<router_id>-route-<destination_cidr>-<next_hop>`: one destination can be served by several routes with different next hops, so only the pair identifies one entry.

## Import

You can import a route by the composite identifier:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_router_route_v2.route_1 <router_id>-route-<destination_cidr>-<next_hop>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/user-types).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `<pool>` — Pool where the router is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `<router_id>` — Unique identifier of the router, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the router ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Cloud routers** tab ⟶ copy the ID of the required router. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack router list` and copy the `ID` field.

* `<destination_cidr>` — Destination of the route, for example, `10.0.0.0/8`.

* `<next_hop>` — Next hop of the route, for example, `192.168.199.10`.

An identifier without the `-route-` separator, or whose remainder does not split into exactly two dash-separated parts, is rejected locally without calling the API.

## Behavior notes

`next_hop` must be an address in a subnet connected to one of the router's internal interfaces. You cannot use an address from the router's external gateway subnet.

Creating a route whose `destination_cidr` and `next_hop` pair already exists fails. Import the existing route if Terraform should manage it.

If the route is removed from the router outside Terraform, or the router itself is deleted, the resource leaves the state on the next refresh without an error and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead.

Destroying the resource removes only its own entry. If the entry is already gone, or the router no longer exists, the destroy succeeds without rewriting the route list.
