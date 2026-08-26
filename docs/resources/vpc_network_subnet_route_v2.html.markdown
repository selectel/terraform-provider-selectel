---
layout: "selectel"
page_title: "Selectel: selectel_vpc_network_subnet_route_v2"
sidebar_current: "docs-selectel-resource-vpc-network-subnet-route-v2"
description: |-
  Creates and manages a single route of a subnet for Selectel products using public API v2.
---

# selectel\_vpc\_network\_subnet\_route\_v2

Creates and manages a single route of a subnet using public API v2. For more information about static routes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/static-routes/).

The resource manages a single entry of the subnet's `host_routes` list and leaves the other entries alone. Use it when different routes of one subnet are owned by different configurations or modules. The subnet itself is managed with the [selectel_vpc_network_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) resource.

~> **Note:** Adding an entry rewrites the whole `host_routes` list of the subnet, so the provider serializes concurrent routes of the same subnet. Do not manage the same subnet's routes both with this resource and by another means at the same time.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_network_subnet_route_v2" "route_1" {
  project_id       = selectel_vpc_project_v2.project_1.id
  region           = "ru-6"
  subnet_id        = selectel_vpc_network_subnet_v2.subnet_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.10"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new route. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the subnet is located, for example, `ru-3`. Changing this creates a new route. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `subnet_id` - (Required) Unique identifier of the subnet the route belongs to. Changing this creates a new route. Retrieved from the [selectel_vpc_network_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) resource.

* `destination_cidr` - (Required) Destination CIDR of the route. Changing this creates a new route. The format is validated by the API rather than by the provider.

* `next_hop` - (Required) Next hop of the route, which must be an address reachable through the subnet. Changing this creates a new route. The format is validated by the API rather than by the provider.

## Attributes Reference

No attributes are exported beyond the arguments. The parent [selectel_vpc_network_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) resource and the corresponding data source expose the full list through `host_routes`.

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

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/user-types).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `<pool>` — Pool where the subnet is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `<subnet_id>` — Unique identifier of the subnet, for example, `8f6c1e9b-2a3d-4f50-9b8a-1c2d3e4f5a6b`. To get the subnet ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ the network page ⟶ the **Subnets** tab ⟶ copy the ID of the required subnet. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack subnet list` and copy the `ID` field.

* `<destination_cidr>` — Destination CIDR of the route, for example, `10.0.0.0/8`.

* `<next_hop>` — Next hop of the route, for example, `192.168.199.10`.

For example:

```shell
terraform import selectel_vpc_network_subnet_route_v2.route_1 8f6c1e9b-2a3d-4f50-9b8a-1c2d3e4f5a6b-route-10.0.0.0/8-192.168.199.10
```

An identifier without the `-route-` separator, or whose remainder does not split into exactly two dash-separated parts, is rejected locally without calling the API.

## Behavior notes

* A route whose `destination_cidr` and `next_hop` pair already exists on the subnet is refused with a diagnostic error. The same `destination_cidr` with a different `next_hop` is a separate route and is not a conflict.

* If the parent subnet is absent when the route is created, creation fails. If the subnet is absent when the route is deleted, deletion succeeds because the route is already gone.

* If the route disappears from `host_routes` outside Terraform, the next refresh removes it from state. Destroying an already absent route succeeds.
