---
layout: "selectel"
page_title: "Selectel: selectel_global_router_vpc_network_v1"
sidebar_current: "docs-selectel-resource-global-router-vpc-network-v1"
description: |-
  Creates and manages a global router network that connects a cloud platform private network to a global router in the Global Router service using public API v1.
---

# selectel\_global\_router\_cloud\_network\_v1

Creates and manages a global router network that connects an existing  cloud platform private network to a global router in the Global Router service using public API v1. To create a cloud platform network, use the [openstack_networking_network_v2](https://docs.selectel.ru/en/terraform/openstack-provider-reference/networking-neutron/resources/openstack_networking_network_v2/) resource. 

For more information about cloud platform private networks and subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/). For more information about global routers, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/).

## Example Usage

```hcl
resource "selectel_global_router_vpc_network_v1" "global_router_vpc_network_1" {
  router_id     = selectel_global_router_router_v1.global_router_1.id
  zone_id       = data.selectel_global_router_zone_v1.zone_1.id
  os_network_id = data.openstack_networking_network_v2.network_1.id
  project_id    = selectel_vpc_project_v2.project_1.id
  name          = "my_super_vpc_net"
  tags          = ["blue", "red"]
}
```

## Argument Reference

* `name` - (Required) Name of the global router network. Does not have to match the name of the cloud platform network.
* `router_id` - (Required) Unique identifier of the global router to which the network will be connected. Retrieved from the [global_router_router_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/global_router_router_v1) resource. Changing this deletes the global router network, connected subnets and static routes and recreates them with the new argument value.
* `zone_id` - (Required) Unique identifier of the zone to which the network will be connected. Retreived from the [selectel_global_router_zone_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/global_router_zone_v1) data source. For cloud platform networks, must be a zone from the `vpc` service. Changing this deletes the global router network, connected subnets and static routes and recreates them with the new argument value.
* `os_network_id` - (Required) Unique identifier of the cloud platform network, retrieved from the [openstack_networking_network_v2](https://docs.selectel.ru/en/terraform/openstack-provider-reference/networking-neutron/data-sources/openstack_networking_network_v2/) data source. Changing this deletes the global router network, connected subnets and static routes and recreates them with the new argument value.
* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). Changing this deletes the global router network, connected subnets and static routes and recreates them with the new argument value.
* `tags` - (Optional) List of global router network tags.

## Attributes Reference

* `id` - Unique identifier of the global router network.
* `name` - Name of the global router network.
* `router_id` - Unique identifier of the global router to which the network is connected.
* `zone_id` - Unique identifier of the zone to which the network is connected.
* `os_network_id` - Unique identifier of the connected cloud platform network.
* `project_id` - Unique identifier of the associated project.
* `tags` - List of global router network tags.
* `vlan` - Network VLAN.
* `created_at` - Time when the global router network was created.
* `updated_at` - Time when the global router network was updated.
* `status` - Global router network status.
* `account_id` - Selectel account ID.
* `netops_vlan_uuid` - Option for internal usage.
* `sv_network_id` - Option for internal usage.
