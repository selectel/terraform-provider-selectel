---
layout: "selectel"
page_title: "Selectel: selectel_global_router_vpc_subnet_v1"
sidebar_current: "docs-selectel-resource-global-router-vpc-subnet-v1"
description: |-
  Creates and manages a global router subnet that connects a cloud platform private subnet to a global router in the Global Router service using public API v1.
---

# selectel\_global\_router\_vpc\_subnet\_v1

Creates and manages a global router subnet that connects an existing cloud platform private subnet to a global router in the Global Router service using public API v1. To create a cloud platform subnet, use the [openstack_networking_subnet_v2](https://docs.selectel.ru/en/terraform/openstack-provider-reference/networking-neutron/resources/openstack_networking_subnet_v2/) resource. 

For more information about cloud platform private networks and subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/). For more information about global routers, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/).

## Example Usage

```hcl
resource "selectel_global_router_vpc_subnet_v1" "global_router_vpc_subnet_1" {
  network_id        = selectel_global_router_vpc_network_v1.global_router_vpc_network_1.id
  os_subnet_id      = data.openstack_networking_subnet_v2.subnet_1.id
  cidr              = "10.10.10.0/24"
  gateway           = "10.10.10.13"
  service_addresses = ["10.10.10.253", "10.10.10.254"]
  name              = "my_super_vpc_subnet"
  tags              = ["blue", "red"]
}
```

## Argument Reference

* `name` - (Required) Name of the global router subnet. Does not have to match the name of the cloud platform subnet.
* `network_id` - (Required) Unique identifier of the global router network that was created for the cloud platform network to which the subnet belongs. Retrieved from the [selectel_global_router_vpc_network_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/global_router_vpc_network_v1) resource. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `cidr` - (Required) Subnet IP address range in CIDR notation. Retrieved from the [openstack_networking_subnet_v2](https://docs.selectel.ru/en/terraform/openstack-provider-reference/networking-neutron/data-sources/openstack_networking_subnet_v2/) data source. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `os_subnet_id` - (Required) Unique identifier of the cloud platform subnet. Retrieved from the [openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_subnet_v2) data source. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `gateway` - (Optional) Subnet IP address that will be used as gateway on the global router. This IP address must be available. If not specified, the first IP address in the subnet range will be used. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `service_addresses` - (Optional) Two subnet IP addresses that will be reserved as service ones. These IP addresses must be available. If not specified, the last two IP addresses in subnet range will be reserved. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.


## Attributes Reference

* `id` - Unique identifier of the global router subnet.
* `name` - Name of the global router subnet.
* `network_id` - Unique identifier of the global router network to which the subnet belongs.
* `cidr` - Subnet IP address range in CIDR notation.
* `os_subnet_id` - Unique identifier of the connected cloud platform subnet.
* `gateway` - Subnet IP address that is used as gateway on the global router.
* `service_addresses` - Two subnet IP addresses that are reserved as service ones.
* `project_id` - Unique identifier of the associated project. 
* `tags` - List of global router subnet tags.
* `created_at` - Time when the global router subnet was created.
* `updated_at` - Time when the global router subnet was updated.
* `status` - Global router subnet status.
* `account_id` - Selectel account ID.
* `netops_subnet_id` - Option for internal usage.
* `sv_subnet_id` - Option for internal usage.
