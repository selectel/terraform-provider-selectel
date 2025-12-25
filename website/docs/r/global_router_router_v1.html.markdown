---
layout: "selectel"
page_title: "Selectel: selectel_global_router_router_v1"
sidebar_current: "docs-selectel-resource-global-router-router-v1"
description: |-
  Creates and manages a global router in the Selectel Global Router service using public API v1.
---

# selectel\_global\_router\_router\_v1

Creates and manages a global router in the Selectel Global Router service using public API v1. For more information about global routers, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/about-global-router/#principle-of-operation).

## Example Usage

```hcl
resource "selectel_global_router_router_v1" "global_router_1" {
  name = "test_router"
  tags = ["blue", "red"]
}
```

## Argument Reference

* `name` - (Required) Name of the router.
* `tags` - (Optional) List of router tags in string format.

## Attributes Reference

* `id` - Unique identifier of the router.
* `name` - Name of the router.
* `tags` - List of router tags.
* `created_at` - Time when the router was created.
* `updated_at` - Time when the router was updated.
* `status` - Router status. For more information, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/router/router-status/).
* `enabled` - Boolean flag, indicates whether the router is active. `False` means that the router is temporarily under maintenance and its connected networks and subnets cannot be updated or deleted, and new networks and subnets cannot be connected.
* `account_id` - Selectel account ID.
* `project_id` - Unique identifier of the associated project.
* `netops_router_id` - Option for internal usage.
* `leak_uuid` - Unique identifier for a group of routers united by a single logical entity.
* `prefix_pool_id` - Unique identifier of prefix pool. Can be used to request a list of allowed subnet prefixes that can be connected to the router.
* `vpn_id` - Option for internal usage.
