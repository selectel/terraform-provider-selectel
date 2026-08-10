---
layout: "selectel"
page_title: "Selectel: selectel_vpc_router_v2"
sidebar_current: "docs-selectel-resource-vpc-router-v2"
description: |-
  Creates and manages a router in the Selectel cloud using Neutron API v2.0.
---

# selectel\_vpc\_router\_v2

Creates and manages a router: it routes traffic between the subnets attached to it and, when an external gateway is set, to and from an external network.

Subnets and ports are attached with [selectel_vpc_router_interface_v2](vpc_router_interface_v2.html), and static routes are managed with [selectel_vpc_router_route_v2](vpc_router_route_v2.html) — this resource does not carry them.

## Example Usage

```hcl
resource "selectel_vpc_router_v2" "router_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-3"
  name        = "router-1"
  description = "Managed by Terraform"
  tags        = ["production", "networking"]

  external_network_id = data.selectel_vpc_network_v2.external.id
  enable_snat         = true
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new router.

* `region` - (Required) Selectel region where the router is created, for example, `ru-3`. Changing this creates a new router.

* `name` - (Optional, Computed) Name of the router.

* `description` - (Optional) Description of the router.

* `admin_state_up` - (Optional, Computed) Administrative state of the router. Defaults to the value the API assigns.

* `external_network_id` - (Optional) Unique identifier of the external network the gateway is attached to. Can be changed without replacement, which moves the gateway; removing it detaches the gateway.

* `enable_snat` - (Optional, Computed) Specifies whether source NAT is enabled on the gateway. Requires `external_network_id` in the same configuration — a value without it is rejected at plan. Can be changed without replacement.

* `flavor_id` - (Optional, Computed) Unique identifier of the router flavor. Changing this creates a new router.

* `availability_zone_hints` - (Optional, Computed) Availability zones the router is requested in. Changing this creates a new router.

* `tags` - (Optional) Tags of the router that are managed by Terraform.

~> **Note:** A caller without the Keystone admin role and without the service role can only use `availability_zone_hints` values from the region's `allowed_router_zones` allowlist. A value outside a non-empty allowlist fails with an authorization error, and an unset allowlist fails with a not-found error — the two are reported as distinct diagnostics, and neither creates the router.

## Attributes Reference

* `status` - Status of the router, for example, `ACTIVE` or `DOWN`.

* `external_fixed_ip` - Gateway addresses assigned by the API:

  * `subnet_id` - Unique identifier of the subnet the address comes from.

  * `ip_address` - Assigned address.

* `availability_zones` - Availability zones the router actually runs in.

* `all_tags` - All tags of the router, including those assigned outside Terraform.

* `blocked` - Specifies whether the router is administratively blocked.

~> **Note:** The gateway addresses cannot be chosen: the cloud's Neutron policy reserves `external_gateway_info:external_fixed_ips` for the Keystone admin role, so the API selects the subnet and the address and this resource reports the result. `qos_policy_id` of the gateway is reserved the same way and is not exposed. `distributed` and `ha` are not exposed at all — the policy reserves reading them for the admin role too, so an ordinary caller never receives them and there would be no actual value to report.

~> **Note:** While `blocked` is `true`, the API refuses update and delete from a caller without the admin role — including the router owner and the holders of `vpc.admin` and `vpc.external_access.admin` — until an admin unblocks the router.

## Behavior notes

A router with attached interfaces is not deleted: the API refuses it with a conflict, and the resource stays in the state — this resource never detaches interfaces on your behalf. Remove the `selectel_vpc_router_interface_v2` resources first.

A router deleted outside Terraform leaves the state on the next refresh without an error, and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead.

## Import

You can import a router:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_vpc_router_v2.router_1 <router_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the router is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<router_id>` — Unique identifier of the router, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`.
