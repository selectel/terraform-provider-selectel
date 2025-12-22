---
layout: "selectel"
page_title: "Selectel: selectel_global_router_vpc_subnet_v1"
sidebar_current: "docs-selectel-resource-global-router-vpc-subnet-v1"
description: |-
  Creates and manages a global router subnet that connects a cloud platform private subnet to a global router in the Selectel Global Router service using public API v1.
---

# selectel\_global\_router\_vpc\_subnet\_v1

Creates and manages a global router subnet that connects a cloud platform private subnet to a global router in the Selectel Global Router service using public API v1. The resource does not create a cloud platform subnet, it must be created before the connection, use the [openstack_networking_subnet_v2](https://docs.selectel.ru/en/terraform/openstack-provider-reference/networking-neutron/resources/openstack_networking_subnet_v2/) resource. 

For more information about cloud platform private networks and subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/). For more information about global router, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/).

## Example Usage

```hcl
resource "selectel_global_router_vpc_subnet_v1" "global_router_vpc_subnet_1" {
  network_id        = "b940567d-439e-4714-ac42-e3f5d4adddf3"
  os_subnet_id      = "92010a80-32ef-45a0-9166-3a3a411e6cd7"
  cidr              = "10.10.10.0/24"
  gateway           = "10.10.10.13"
  service_addresses = ["10.10.10.253", "10.10.10.254"]
  name              = "my_super_vpc_subnet"
  tags              = ["blue", "red"]
}
```

## Argument Reference

* `name` - (Required) Name of the global router subnet. Does not have to match the name of the cloud platform subnet.
* `network_id` - (Required) Unique identifier of the global router network that was created for the cloud platform network the subnet belongs to. Retrieved from the [selectel_global_router_vpc_network_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/global_router_vpc_network_v1) resource. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `cidr` - (Required) Subnet IP address range in CIDR notation. Retrieved from the [openstack_networking_subnet_v2](https://docs.selectel.ru/en/terraform/openstack-provider-reference/networking-neutron/data-sources/openstack_networking_subnet_v2/) data source. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `os_subnet_id` - (Required) Unique identifier of the cloud platform subnet. Retrieved from the [openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_subnet_v2) data source. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `gateway` - (Optional) Subnet IP address that will be used as gateway on the global router. This IP address must be available. If not specified, the first IP address in subnet range will be used. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `service_addresses` - (Optional) Two of the subnet IP addresses that will be reserved as service ones. These IP addresses must be available. If not specified, the last two IP addresses in subnet range will be reserved. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.


## Attributes Reference

* `id` - Unique identifier of the global router subnet.
* `name` - Name of the global router subnet.
* `network_id` - Unique identifier of the global router network the subnet belongs to.
* `cidr` - Subnet IP address range in CIDR notation.
* `os_subnet_id` - Unique identifier of the connected cloud platform subnet.
* `gateway` - Subnet IP address that is used as gateway on the global router.
* `service_addresses` - Two of the subnet IP addresses that are reserved as service ones.
* `project_id` - Unique identifier of the associated project. 
* `tags` - List of global router subnet tags.
* `created_at` - Time when the global router subnet was created.
* `updated_at` - Time when the global router subnet was updated.
* `status` - Global router subnet status.
* `account_id` - Selectel account ID.
* `netops_subnet_id` - Option for internal usage.
* `sv_subnet_id` - Option for internal usage.
