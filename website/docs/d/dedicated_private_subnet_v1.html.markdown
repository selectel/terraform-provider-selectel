---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_private_subnet_v1"
sidebar_current: "docs-selectel-data-source-dedicated-private-subnet-v1"
description: |-
  Provides a list of existing private subnets in a project. 
---

# selectel\_dedicated\_private\_subnet\_v1

Provides a list of existing private subnets in a project. For more information about private subnets for dedicated servers, see the [official Selectel documentation.](https://docs.selectel.ru/en/dedicated/networks/private-networks-and-subnets/#assign-private-ip-address-to-server)

## Example usage

```hcl
data "selectel_dedicated_private_subnet_v1" "subnet_ds" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    location_id = data.selectel_dedicated_location_v1.server_location.locations[0].id
    vlan        = "1000"
    subnet      = "192.168.100.0/24"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `filter` - (Optional) Filter for searching subnets.
  * `location_id` - (Required) Unique identifier of the location where the private subnet will be created. Retrieved from the [dedicated_location_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/dedicated_location_v1) data source.
  * `subnet` - (Optional) Subnet in CIDR notation, for example, 192.168.100.0/24. 
  * `vlan` - (Optional) Unique identifier of VLAN.
  * `ip` - (Optional) IP address belonging to the subnet.

## Attributes Reference

* `subnets` - List of available subnets:
  * `id` - Unique identifier of the subnet (UUID).
  * `subnet` - Subnet in CIDR notation, for example, 192.168.100.0/24.
  * `vlan` - Unique identifier of the VLAN.
  * `reserved_ips` - List of reserved IP addresses in the subnet that you cannot use.

