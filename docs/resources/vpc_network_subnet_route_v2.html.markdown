---
layout: "selectel"
page_title: "Selectel: selectel_vpc_network_subnet_route_v2"
sidebar_current: "docs-selectel-resource-vpc-network-subnet-route-v2"
description: |-
  Creates and manages a single route of a subnet in the Selectel cloud using VPC API v2.0.
---

# selectel\_vpc\_network\_subnet\_route\_v2

Creates and manages a single entry of a subnet's `host_routes` list, without touching the other entries. Use it when different routes of one subnet are owned by different configurations or modules.

~> **Note:** Adding an entry rewrites the whole `host_routes` list of the subnet, so the provider serializes concurrent routes of the same subnet. Do not manage the same subnet's routes both with this resource and by another means at the same time.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_network_subnet_route_v2" "route_1" {
  project_id       = selectel_vpc_project_v2.project_1.id
  region           = "ru-3"
  subnet_id        = selectel_vpc_network_subnet_v2.subnet_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.10"
}
```

## Argument Reference

Every argument is required, and changing any of them creates a new route: the entry has no update path of its own.

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new route.

* `region` - (Required) Pool where the subnet is located, for example, `ru-3`. Changing this creates a new route.

* `subnet_id` - (Required) Unique identifier of the subnet the route belongs to. Changing this creates a new route.

* `destination_cidr` - (Required) Destination CIDR of the route. Changing this creates a new route. The format is validated by the API rather than by the provider.

* `next_hop` - (Required) Next hop of the route, which must be an address reachable through the subnet. Changing this creates a new route. The format is validated by the API rather than by the provider.

## Attributes Reference

No attributes are exported beyond the arguments. To read the full list of routes of a subnet, use the [selectel_vpc_network_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/vpc_network_subnet_v2) data source, which exports `host_routes`.

## Behavior notes

* A route whose `destination_cidr` and `next_hop` pair already exists on the subnet is refused with a diagnostic error. The same `destination_cidr` with a different `next_hop` is a separate route and is not a conflict.

* If the parent subnet is already absent when the route is created, no route is created and no error is reported. If the parent subnet is absent when the route is deleted, the delete succeeds, because removing the subnet has already removed the route.

* If the parent subnet still exists but the route has disappeared from its `host_routes` outside Terraform, both read and delete report a diagnostic error rather than silently dropping the resource from the state.

## Import

You can import a route:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_network_subnet_route_v2.route_1 <subnet_id>-route-<destination_cidr>-<next_hop>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<pool>` — Pool where the subnet is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<subnet_id>` — Unique identifier of the subnet, for example, `8f6c1e9b-2a3d-4f50-9b8a-1c2d3e4f5a6b`.

* `<destination_cidr>` — Destination CIDR of the route, for example, `10.0.0.0/8`.

* `<next_hop>` — Next hop of the route, for example, `192.168.199.10`.

For example:

```shell
terraform import selectel_vpc_network_subnet_route_v2.route_1 8f6c1e9b-2a3d-4f50-9b8a-1c2d3e4f5a6b-route-10.0.0.0/8-192.168.199.10
```

An identifier without the `-route-` separator, or whose remainder does not split into exactly two dash-separated parts, is rejected locally without calling the API.
