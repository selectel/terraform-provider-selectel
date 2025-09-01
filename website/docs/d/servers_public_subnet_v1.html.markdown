---
layout: "selectel"
page_title: "Selectel: selectel_servers_public_subnet_v1"
sidebar_current: "docs-selectel-datasource-servers-public-subnet-v1"
description: |-
  Provides a list of available public subnets.
---

# selectel\_servers\_public\_subnet\_v1

Provides a list of available public subnets.

## Example Usage

```hcl
data "selectel_servers_location_v1" "server_locations" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name = "some-subnet"
  }
}

data "selectel_servers_public_subnet_v1" "public_subnets" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    location_id = data.selectel_servers_location_v1.server_locations.locations[0].id
    subnet = "192.168.1.0/29"
    ip = "192.168.1.3"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `filter` - (Required) Values to filter available subnets.

  * `ip` - (Optional) IP address to search included in a subnet.
  * `subnet` - (Optional) Subnet in CIDR notation to search.
  * `location_id` - (Required) Unique identifier of the location.

## Attributes Reference

* `subnets` - List of the available subnets:

  * `id` - Unique identifier of the subnet.
  * `network_id` - Unique identifier of the network.
  * `subnet` - Subnet in CIDR notation.
  * `broadcast` - Broadcast address.
  * `gateway` - Gateway address.
  * `reserved_vrp_ips` - List of reserved VRP IPs.
