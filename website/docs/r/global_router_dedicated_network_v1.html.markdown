---
layout: "selectel"
page_title: "Selectel: selectel_global_router_dedicated_network_v1"
sidebar_current: "docs-selectel-resource-global-router-dedicated-network-v1"
description: |-
  Creates and manages a global router network that connects a dedicated server private network to a global router in the Selectel Global Router service using public API v1.
---

# selectel\_global\_router\_dedicated\_network\_v1

Creates and manages a global router network that connects a dedicated server private network (VLAN) to a global router in the Selectel Global Router service using public API v1. The resourse does not create a network, a private VLAN must be added in the Control Panel before the connection.

For more information about dedicated server networks, see the [official Selectel documentation](https://docs.selectel.ru/en/dedicated/networks/about-networks/). For more information about global router, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/).

## Example Usage

```hcl
resource "selectel_global_router_dedicated_network_v1" "global_router_dedicated_network_1" {
  router_id = "2072eda5-34fe-4a14-80cc-68b472aa9dbf"
  zone_id   = "2b4f3050-6e2d-4563-a287-6c0c1d4ceb3a"
  vlan      = "1234"
  name      = "my_super_dedicated_net"
  tags      = ["blue", "red"]
}
```

## Argument Reference

* `name` - (Required) Name of the global router network.
* `router_id` - (Required) Unique identifier of the global router the network will be connected to. Retrieved from the [global_router_router_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/global_router_router_v1) resource. Changing this deletes the global router network, connected subnets and static routes and recreates them with the new argument value.
* `zone_id` - (Required) Unique identifier of the zone the network will be connected to. Retrieved from the [selectel_global_router_zone_v1](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/global_router_zone_v1) data source. 
    For dedicated server networks, must be a zone from the `dedicated` service. Changing this deletes the global router network, connected subnets and static routes and recreates them with the new argument value.
* `vlan` - (Required) Private VLAN number. To get VLAN number, in the [Control panel](https://my.selectel.ru/servers/network/networks), go to **Dedicated servers** ⟶ the **VLAN** tab ⟶ copy the VLAN number. Changing this deletes the global router network, connected subnets and static routes and recreates them with the new argument value.
* `tags` - (Optional) List of global router network tags in string format.

## Attributes Reference

* `id` - Unique identifier of the global router network.
* `name` - Name of the global router network.
* `router_id` - Unique identifier of the global router the network is connected to.
* `zone_id` - Unique identifier of the zone the network is connected to.
* `vlan` - Network VLAN. 
* `tags` - List of global router network tags.
* `created_at` - Time when the global router network was created.
* `updated_at` - Time when the global router network was updated.
* `status` - Global router network status.
* `account_id` - Selectel account ID.
* `netops_vlan_uuid` - Option for internal usage.
* `sv_network_id` - Option for internal usage.
