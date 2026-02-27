---
layout: "selectel"
page_title: "Selectel: selectel_global_router_static_route_v1"
sidebar_current: "docs-selectel-resource-global-router-static-route-v1"
description: |-
  Creates and manages a static route in the Global Router service using public API v1.
---

# selectel\_global\_router\_static\_route\_v1

Creates and manages a static route in the Global Router service using public API v1. For more information about global router static routes, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/create-network/create-global-router-network/#configure-routing-on-global-router).

~> Note: Next hop IP address in a static route must belong to one of the subnets already connected to the router. If you create the `selectel_global_router_vpc_subnet_v1` resource in the same run as the static route, make sure to use `depends_on` to enforce the global router subnet resource creation before the static route creation is triggered. We strongly recommend to use the `lifecycle` argument that triggers static route recreation in case the parent subnet was recreated.

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

resource "selectel_global_router_static_route_v1" "global_router_static_route_1" {
  router_id = selectel_global_router_router_v1.global_router_1.id
  cidr      = "0.0.0.0/0"
  next_hop  = "10.10.10.42"
  name      = "stat_route_to_dc"
  tags      = ["blue", "red"]

  depends_on = ["selectel_global_router_vpc_subnet_v1.global_router_vpc_subnet_1"]
  lifecycle {
    replace_triggered_by = [
      selectel_global_router_vpc_subnet_v1.global_router_vpc_subnet_1.id
    ]
  }
}
```

## Argument Reference

* `name` - (Required) Name of the static route.
* `router_id` - (Required) Unique identifier of the global router the static route will be created on. Retrieved from the [global_router_router_v1](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/global_router_router_v1) resource. Changing this deletes the static route and recreates it with the new argument value.
* `next_hop` - (Required) IP address in a subnet through which traffic will be routed to the destination subnet. The IP address must belong to one of the subnets connected to the router. Changing this deletes the static route and recreates it with the new argument value.
* `cidr` - (Required) Destination subnet IP address range in CIDR notation to which you direct traffic. Changing this deletes the static route and recreates it with the new argument value.
* `tags` - (Optional) List of static route tags.

## Attributes Reference

* `id` - Unique identifier of the static route.
* `name` - Name of the static route.
* `router_id` - Unique identifier of the router where static route will be created.
* `next_hop` - IP address in a subnet through which traffic is routed to the destination subnet.
* `cidr` - Destination subnet IP address range in CIDR notation to which traffic is redirected.
* `tags` - List of static route tags.
* `created_at` - Time when the static route was created.
* `updated_at` - Time when the static route was updated.
* `status` - Static route status.
* `account_id` - Selectel account ID.
* `project_id` - Unique identifier of the associated project.
* `netops_static_route_id` - Option for internal usage.
* `subnet_id` - Unique identifier of the global router subnet which contains the next hop IP address.

## Import {#import}

You can import a global router static route:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_global_router_static_route_v1.global_router_static_route_1 <static_route_id>
```

where:

*   `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

*   `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/access-management/).

*   `<password>` — Password of the service user.

*   `<static_route_id>` — Unique identifier of the global router static route, for example, `223ddf21-82ca-44a7-9782-88ff29b7d3e4`. To get the global router static route ID in the [Control panel](https://my.selectel.ru/network/localnetwork/l3/), go to **Products** ⟶ **Global Router** ⟶ the global router page ⟶ the **Static routes** tab. The global router static route ID is in the **UUID** column.