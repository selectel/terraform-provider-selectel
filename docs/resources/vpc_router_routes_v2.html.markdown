---
layout: "selectel"
page_title: "Selectel: selectel_vpc_router_routes_v2"
sidebar_current: "docs-selectel-resource-vpc-router-routes-v2"
description: |-
  Creates and manages the complete set of static routes of a cloud router for Selectel products using public API v2.
---

# selectel\_vpc\_router\_routes\_v2

Creates and manages the complete set of static routes of a cloud router using public API v2. For more information about static routes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/static-routes/).

Terraform owns the router's route list: a route added by anything else — a `selectel_vpc_router_route_v2` resource, a direct API call, the Control panel — is reported as drift and removed by the next apply. Destroying the resource removes only the routes recorded in its state; anything else on the router stays. Use [selectel_vpc_router_route_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_route_v2) instead when you want to manage individual routes and leave the rest of the list alone. Do not manage the same router with both resources — they will fight over the list.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_router_routes_v2" "routes_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  router_id  = selectel_vpc_router_v2.router_1.id
  depends_on = [selectel_vpc_router_interface_v2.interface_1]

  routes {
    destination_cidr = "10.0.0.0/8"
    next_hop         = "192.168.199.10"
  }

  routes {
    destination_cidr = "172.16.0.0/12"
    next_hop         = "192.168.199.10"
  }
}
```

~> **Note:** A next hop is only accepted once the router has an interface in the subnet the address belongs to, so depend on the `selectel_vpc_router_interface_v2` resource that attaches it.

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new resource. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the router is located, for example, `ru-3`. Changing this creates a new resource. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `router_id` - (Required) Unique identifier of the router whose routes are managed. Changing this creates a new resource. Retrieved from the [selectel_vpc_router_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2) resource.

* `routes` - (Optional) The complete set of static routes of the router. Omit all `routes` blocks to declare an empty set. The first apply is additive: it only ensures the declared routes are present, so an existing route is adopted rather than deleted; from the next apply on, the declared set is applied as the router's whole route list. The API validates the submitted list as a whole, so one invalid next hop rejects every route in that apply. Each block takes:

  * `destination_cidr` - (Required) Destination of the route, for example, `10.0.0.0/8`.

  * `next_hop` - (Required) Address traffic for `destination_cidr` is forwarded to. Must be an address in a subnet connected to one of the router's internal interfaces — you cannot use an address from the router's external gateway subnet.

~> **Note:** If an apply was rejected by the API, run a plan or an apply before destroying. The rejected list stays in state until a refresh replaces it with the router's actual one, so a destroy in between removes only the rejected entries and leaves the routes the router really has — and those routes keep the router's interfaces from being deleted until they are removed.

## Import

You can import the route set by the identifier of its router:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_router_routes_v2.routes_1 <router_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/user-types).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `<pool>` — Pool where the router is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `<router_id>` — Unique identifier of the router, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the router ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Cloud routers** tab ⟶ copy the ID of the required router. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack router list` and copy the `ID` field.

The import fills `routes` with the router's full current list. A `router_id` that matches no router fails with a diagnostic error instead of adding an empty resource to the state.
