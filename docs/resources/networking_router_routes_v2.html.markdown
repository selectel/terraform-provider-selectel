---
layout: "selectel"
page_title: "Selectel: selectel_networking_router_routes_v2"
sidebar_current: "docs-selectel-resource-networking-router-routes-v2"
description: |-
  Manages the complete set of static routes of a router in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_router\_routes\_v2

Manages the **complete** set of static routes of a router. Terraform owns the router's route list: a route added by anything else is reported as drift and removed by the next apply.

Use [selectel_networking_router_route_v2](networking_router_route_v2.html) instead when you want to manage individual routes and leave the rest of the list alone. Do not manage the same router with both resources — they will fight over the list.

## Example Usage

```hcl
resource "selectel_networking_router_routes_v2" "routes_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  router_id  = selectel_networking_router_v2.router_1.id

  routes {
    destination_cidr = "10.0.0.0/8"
    next_hop         = "192.168.199.10"
  }

  routes {
    destination_cidr = "172.16.0.0/12"
    next_hop         = "192.168.199.10"
  }

  depends_on = [selectel_networking_router_interface_v2.interface_1]
}
```

~> **Note:** A next hop is only accepted once the router has an interface in the subnet the address belongs to, so depend on the `selectel_networking_router_interface_v2` resource that attaches it.

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new resource.

* `region` - (Required) Selectel region where the router is located, for example, `ru-3`. Changing this creates a new resource.

* `router_id` - (Required) Unique identifier of the router whose routes are managed. Changing this creates a new resource.

* `routes` - (Optional, Computed) The complete set of static routes of the router. Omit it, or declare it empty, to state that the router should carry no routes. Each block takes:

  * `destination_cidr` - (Required) Destination of the route, for example, `10.0.0.0/8`.

  * `next_hop` - (Required) Address traffic for `destination_cidr` is forwarded to.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is the identifier of the router.

## Behavior notes

`routes` is authoritative. Because a refresh fills it with the router's actual list, a route added by another mechanism — a `selectel_networking_router_route_v2` resource, a direct API call, the control panel — shows up in the plan as a route to be removed, and the next apply removes it. This is the deliberate difference from the single-route resource, which never touches routes it did not add.

The first apply is additive: it only ensures the declared routes are present, so an existing route is adopted rather than deleted. From the next apply on, the declared list is applied as the router's whole list.

Every `next_hop` must be an address in a subnet served by one of the router's **internal** interfaces. The Selectel Neutron fork excludes the router's ports on external networks when it builds the set of acceptable next hops, so an address inside the router's external gateway subnet is refused — even though upstream Neutron would accept it. The API validates the submitted list as a whole, so one unacceptable next hop rejects every route in that apply and the router keeps the list it already had.

Destroying the resource removes only the routes recorded in its state; anything else on the router stays. If the state holds no routes, the destroy makes no API call. If the router is already gone, or the managed routes are already absent, the destroy succeeds without an error.

## Import

You can import the route set by the identifier of its router:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_networking_router_routes_v2.routes_1 <router_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the router is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<router_id>` — Unique identifier of the router, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`.

The import fills `routes` with the router's full current list. A `router_id` that matches no router fails with a diagnostic error instead of adding an empty resource to the state.
