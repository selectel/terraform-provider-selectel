---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_public_subnet_v1"
sidebar_current: "docs-selectel-datasource-dedicated-public-subnet-v1"
description: |-
  Provides a list of available public subnets.
---

# selectel\_dedicated\_public\_subnet\_v1

Provides a list of available additional public subnets. Learn more about [Additional public subnets](https://docs.selectel.ru/en/dedicated/networks/ip-addresses/#add-additional-ips)

## Example Usage

```hcl
data "selectel_dedicated_public_subnet_v1" "public_subnets" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    location_id = data.selectel_dedicated_location_v1.server_location.locations[0].id
    subnet = "192.168.1.0/29"
    ip = "192.168.1.3"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `filter` - (Optional) Values to filter available subnets.

  * `ip` - (Optional) IP address to search included in a subnet.
  * `subnet` - (Optional) Subnet in CIDR notation to search.
  * `location_id` - (Optional) Unique identifier of the location. Retrieved from the [selectel_dedicated_location_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dedicated_location_v1) data source. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/availability-matrix/#dedicated-servers).

## Attributes Reference

* `subnets` - List of the available subnets:

  * `id` - Unique identifier of the subnet.
  * `network_id` - Unique identifier of the network.
  * `subnet` - Subnet in CIDR notation.
  * `broadcast` - Broadcast address.
  * `gateway` - Gateway address.
  * `reserved_vrrp_ips` - List of reserved VRRP IPs.
  * `ip` - IP address from the filter. Can be used to pass forward.
