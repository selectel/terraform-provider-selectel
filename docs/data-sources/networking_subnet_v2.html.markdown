---
layout: "selectel"
page_title: "Selectel: selectel_networking_subnet_v2"
sidebar_current: "docs-selectel-datasource-networking-subnet-v2"
description: |-
  Provides information about a subnet in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_subnet\_v2

Provides information about a subnet. For more information about networks and subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/servers/networks/about-networks/).

The data source returns exactly one subnet. If no subnet matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

## Example Usage

```hcl
data "selectel_networking_subnet_v2" "subnet_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  cidr       = "192.168.199.0/24"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.

* `region` - (Required) Pool where the subnet is located, for example, `ru-3`.

You can combine any of the following selectors. The subnet must match all of them at the same time:

* `subnet_id` - (Optional) Unique identifier of the subnet.

* `name` - (Optional) Name of the subnet.

* `description` - (Optional) Description of the subnet.

* `network_id` - (Optional) Unique identifier of the parent network.

* `cidr` - (Optional) CIDR of the subnet.

* `gateway_ip` - (Optional) Gateway of the subnet.

* `ip_version` - (Optional) Internet protocol version: `4` or `6`.

* `dhcp_enabled` - (Optional) Specifies whether DHCP is enabled for the subnet.

* `ipv6_address_mode` - (Optional) IPv6 address assignment mode: `slaac`, `dhcpv6-stateful`, or `dhcpv6-stateless`.

* `ipv6_ra_mode` - (Optional) IPv6 router advertisement mode, with the same values as `ipv6_address_mode`.

* `subnetpool_id` - (Optional) Unique identifier of the subnet pool the CIDR was allocated from.

* `tags` - (Optional) Tags that the subnet must have.

## Attributes Reference

* `enable_dhcp` - Specifies whether DHCP is enabled for the subnet.

* `dns_nameservers` - DNS name servers of the subnet.

* `dns_publish_fixed_ip` - Specifies whether the fixed IPs of the subnet are published to the DNS service.

* `allocation_pools` - Address ranges available for allocation:

  * `start` - First address of the pool.

  * `end` - Last address of the pool.

* `host_routes` - All routes of the subnet, including the ones added by the [selectel_networking_subnet_route_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/networking_subnet_route_v2) resource:

  * `destination_cidr` - Destination CIDR of the route.

  * `next_hop` - Next hop of the route.

* `service_types` - Service types associated with the subnet.

* `all_tags` - All tags of the subnet.

* `blocked` - Specifies whether the subnet is administratively blocked.

~> **Note:** The returned attributes report the actual state of the subnet and cannot be used to change it. `segment_id` is not exposed: the cloud's Neutron policy reserves it for the Keystone admin role and the networking SDK does not return it.
