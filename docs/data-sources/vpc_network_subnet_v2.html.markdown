---
layout: "selectel"
page_title: "Selectel: selectel_vpc_network_subnet_v2"
sidebar_current: "docs-selectel-datasource-vpc-network-subnet-v2"
description: |-
  Provides information about a subnet of a private network in Selectel products.
---

# selectel\_vpc\_network\_subnet\_v2

Provides information about a subnet of a private network. For more information about private networks and subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/).

The data source returns exactly one subnet. You can combine any of the optional selectors below, and the subnet must match all of them at the same time. If no subnet matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to replace the corresponding OpenStack data source with the Selectel provider data source.

## Example Usage

```hcl
data "selectel_vpc_network_subnet_v2" "subnet_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  cidr       = "192.168.199.0/24"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the subnet is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `subnet_id` - (Optional) Unique identifier of the subnet. To get the subnet ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ the network page ⟶ the **Subnets** tab ⟶ copy the ID of the required subnet. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack subnet list` and copy the `ID` field.

* `name` - (Optional) Name of the subnet.

* `description` - (Optional) Description of the subnet.

* `network_id` - (Optional) Unique identifier of the parent network. To get the network ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ copy the ID of the required network. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack network list` and copy the `ID` field.

* `cidr` - (Optional) CIDR of the subnet.

* `gateway_ip` - (Optional) Gateway of the subnet.

* `ip_version` - (Optional) Internet protocol version: `4` or `6`.

* `enable_dhcp` - (Optional) Specifies whether DHCP is enabled for the subnet.

* `ipv6_address_mode` - (Optional) IPv6 address assignment mode: `slaac`, `dhcpv6-stateful`, or `dhcpv6-stateless`.

* `ipv6_ra_mode` - (Optional) IPv6 router advertisement mode, with the same values as `ipv6_address_mode`.

* `tags` - (Optional) Tags that the subnet must have.

## Attributes Reference

* `dns_nameservers` - Ordered list of DNS name servers passed to instances through DHCP. This setting does not create Private DNS records.

* `allocation_pool` - Unordered set of address ranges available for allocation:

  * `start` - First address of the pool.

  * `end` - Last address of the pool.

* `host_routes` - All routes of the subnet, including the ones added by the [selectel_vpc_network_subnet_route_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_route_v2) resource:

  * `destination_cidr` - Destination CIDR of the route.

  * `next_hop` - Next hop of the route.

* `service_types` - Unordered set of service types associated with the subnet.

* `all_tags` - All tags of the subnet.

* `blocked` - Specifies whether the subnet is administratively blocked.
