---
layout: "selectel"
page_title: "Selectel: selectel_networking_router_route_v2"
sidebar_current: "docs-selectel-resource-networking-router-route-v2"
description: |-
  Manages one static route of a router in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_router\_route\_v2

Manages exactly one static route of a router — one `destination_cidr` and `next_hop` pair — leaving every other route of that router alone.

Use [selectel_networking_router_routes_v2](networking_router_routes_v2.html) instead when you want Terraform to own the router's complete route list.

## Example Usage

```hcl
resource "selectel_networking_router_route_v2" "route_1" {
  project_id       = selectel_vpc_project_v2.project_1.id
  region           = "ru-3"
  router_id        = selectel_networking_router_v2.router_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.10"

  depends_on = [selectel_networking_router_interface_v2.interface_1]
}
```

~> **Note:** A next hop is only accepted once the router has an interface in the subnet the address belongs to, so depend on the `selectel_networking_router_interface_v2` resource that attaches it.

## Argument Reference

Every argument is required, and changing any of them creates a new route — the resource has no update path of its own.

* `project_id` - (Required) Unique identifier of the associated project.

* `region` - (Required) Selectel region where the router is located, for example, `ru-3`.

* `router_id` - (Required) Unique identifier of the router the route belongs to.

* `destination_cidr` - (Required) Destination of the route, for example, `10.0.0.0/8`.

* `next_hop` - (Required) Address traffic for `destination_cidr` is forwarded to.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is `<router_id>-route-<destination_cidr>-<next_hop>`: one destination can be served by several routes with different next hops, so only the pair identifies one entry.

## Behavior notes

`next_hop` must be an address in a subnet served by one of the router's **internal** interfaces. The Selectel Neutron fork excludes the router's ports on external networks when it builds the set of acceptable next hops, so an address inside the router's external gateway subnet is refused — even though upstream Neutron, and the equivalent resource of terraform-provider-openstack against vanilla Neutron, would accept it. Both cases are reported as the same class of error, from the API, and neither changes the router.

Applying a route that already exists on the router with exactly the same pair — for example one added outside Terraform — succeeds and adopts that entry, without rewriting the router's route list and without a conflict error.

If the route is removed from the router outside Terraform, or the router itself is deleted, the resource leaves the state on the next refresh without an error and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead.

Destroying the resource removes only its own entry. If the entry is already gone, or the router no longer exists, the destroy succeeds without rewriting the route list.

## Import

You can import a route by the composite identifier:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_networking_router_route_v2.route_1 <router_id>-route-<destination_cidr>-<next_hop>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the router is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<router_id>` — Unique identifier of the router, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`.

* `<destination_cidr>` — Destination of the route, for example, `10.0.0.0/8`.

* `<next_hop>` — Next hop of the route, for example, `192.168.199.10`.

An identifier without the `-route-` separator, or whose remainder does not split into exactly two dash-separated parts, is rejected locally without calling the API.
