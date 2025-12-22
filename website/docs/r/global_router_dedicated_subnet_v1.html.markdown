---
layout: "selectel"
page_title: "Selectel: selectel_global_router_dedicated_subnet_v1"
sidebar_current: "docs-selectel-resource-global-router-dedicated-subnet-v1"
description: |-
  Creates and manages a global router subnet that connects a dedicated server private subnet to a global router in the Selectel Global Router service using public API v1.
---

# selectel\_global\_router\_dedicated\_subnet\_v1

Creates and manages a global router subnet that connects a dedicated server private subnet to a global router in the Selectel Global Router service using public API v1. The resourсe does not create a dedicated server subnet, it must be added in the Control Panel before the connection. Learn how to [add a private subnet in the control panel](https://docs.selectel.ru/en/dedicated/networks/ip-addresses/#add-private-subnet-to-control-panel).

For more information about dedicated server networks, see the [official Selectel documentation](https://docs.selectel.ru/en/dedicated/networks/about-networks/). For more information about global router, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/).


## Example Usage

```hcl
resource "selectel_global_router_dedicated_subnet_v1" "global_router_dedicated_subnet_1" {
  network_id        = "097409ac-d4b1-4709-afd6-58a4747f1586"
  cidr              = "10.10.10.0/24"
  gateway           = "10.10.10.13"
  service_addresses = ["10.10.10.253", "10.10.10.254"]
  name              = "my_super_dedicated_subnet"
  tags              = ["blue", "red"]
}
```

## Argument Reference

* `name` - (Required) Name of the global router subnet.
* `network_id` - (Required) Unique identifier of the global router network, that was created for the dedicated server network the subnet belongs to. Retrieved from the [selectel_global_router_dedicated_network_v1](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/global_router_dedicated_network_v1) resource.  Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `cidr` - (Required) Subnet IP address range in CIDR notation. To get subnet CIDR, in the [Control panel](https://my.selectel.ru/servers/network/networks), go to **Dedicated servers** ⟶ the **Private subnets** tab ⟶ copy the subnet CIDR. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `gateway` - (Optional) Subnet IP address that will be used as gateway on the global router. This IP address must be available. If not specified, the first IP address in subnet range will be used. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `service_addresses` - (Optional) Two of the subnet IP addresses that will be reserved as service ones. These IP addresses must be available. If not specified, the last two IP addresses in subnet range will be reserved. Changing this deletes the global router subnet and connected static routes and recreates them with the new argument value.
* `tags` — (Optional) List of global router subnet tags in string format.
## Attributes Reference

* `id` - Unique identifier of the global router subnet.
* `name` - Name of the global router subnet.
* `network_id` - Unique identifier of the global router network the subnet belongs to.
* `cidr` - Subnet IP address range in CIDR notation.
* `gateway` - Subnet IP address that is used as gateway on the global router.
* `service_addresses` - Two of the subnet IP addresses that are reserved as service ones.
* `tags` - List of subnet tags.
* `created_at` - Time when the global router subnet was created.
* `updated_at` - Time when the global router subnet was updated.
* `status` - Global router subnet status.
* `account_id` - Selectel account ID.
* `netops_subnet_id` - Option for internal usage.
* `sv_subnet_id` - Option for internal usage.
