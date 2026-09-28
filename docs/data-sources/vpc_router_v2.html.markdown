---
layout: "selectel"
page_title: "Selectel: selectel_vpc_router_v2"
sidebar_current: "docs-selectel-datasource-vpc-router-v2"
description: |-
  Provides information about a cloud router in Selectel products.
---

# selectel\_vpc\_router\_v2

Provides information about a cloud router. For more information about cloud routers, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/cloud-routers/).

The data source returns exactly one router. You can combine any of the optional selectors below, and the router must match all of them at the same time. If no router matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to replace the corresponding OpenStack data source with the Selectel provider data source.

## Example Usage

```hcl
data "selectel_vpc_router_v2" "router_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name       = "router-1"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the router is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `router_id` - (Optional) Unique identifier of the router.

* `name` - (Optional) Name of the router.

* `description` - (Optional) Description of the router.

* `admin_state_up` - (Optional) Administrative state of the router.

* `tags` - (Optional) Tags that the router must have.

## Attributes Reference

* `status` - Status of the router, for example, `ACTIVE` or `DOWN`.

* `external_network_id` - Unique identifier of the external network the gateway is attached to, empty when the router has no gateway.

* `enable_snat` - Specifies whether source NAT is enabled on the gateway.

* `external_fixed_ip` - Gateway addresses assigned by the API:

  * `subnet_id` - Unique identifier of the subnet the address comes from.

  * `ip_address` - Assigned address.

* `routes` - Static routes of the router:

  * `destination_cidr` - Destination of the route.

  * `next_hop` - Next hop of the route.

* `availability_zone_hints` - Pool segments the router was requested in.

* `availability_zones` - Pool segments the router actually runs in.

* `all_tags` - All tags of the router.

* `blocked` - Specifies whether the router is administratively blocked.

~> **Note:** `routes` reports every static route of the router, including routes added by [selectel_vpc_router_route_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_route_v2) and [selectel_vpc_router_routes_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_routes_v2).
