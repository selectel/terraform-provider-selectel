---
layout: "selectel"
page_title: "Selectel: selectel_vpc_router_v2"
sidebar_current: "docs-selectel-datasource-vpc-router-v2"
description: |-
  Provides information about a router in the Selectel cloud using VPC API v2.0.
---

# selectel\_vpc\_router\_v2

Provides information about a router.

The data source returns exactly one router. If no router matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html#migrating-data-sources) to replace the corresponding OpenStack data source with the Selectel provider data source.

## Example Usage

```hcl
data "selectel_vpc_router_v2" "router_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "router-1"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.

* `region` - (Required) Selectel region where the router is located, for example, `ru-3`.

You can combine any of the following selectors. The router must match all of them at the same time:

* `router_id` - (Optional, Computed) Unique identifier of the router.

* `name` - (Optional, Computed) Name of the router.

* `description` - (Optional, Computed) Description of the router.

* `admin_state_up` - (Optional, Computed) Administrative state of the router.

* `status` - (Optional, Computed) Status of the router, for example, `ACTIVE` or `DOWN`.

* `tags` - (Optional) Tags that the router must have.

## Attributes Reference

* `external_network_id` - Unique identifier of the external network the gateway is attached to, empty when the router has no gateway.

* `enable_snat` - Specifies whether source NAT is enabled on the gateway.

* `external_fixed_ip` - Gateway addresses assigned by the API:

  * `subnet_id` - Unique identifier of the subnet the address comes from.

  * `ip_address` - Assigned address.

* `routes` - Static routes of the router:

  * `destination_cidr` - Destination of the route.

  * `next_hop` - Next hop of the route.

* `flavor_id` - Unique identifier of the router flavor.

* `availability_zone_hints` - Availability zones the router was requested in.

* `availability_zones` - Availability zones the router actually runs in.

* `all_tags` - All tags of the router.

* `blocked` - Specifies whether the router is administratively blocked.

~> **Note:** `routes` reports every static route of the router, including routes added by [selectel_vpc_router_route_v2](../r/vpc_router_route_v2.html) and [selectel_vpc_router_routes_v2](../r/vpc_router_routes_v2.html).
