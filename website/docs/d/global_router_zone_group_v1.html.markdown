---
layout: "selectel"
page_title: "Selectel: selectel_global_router_zone_group_v1"
sidebar_current: "docs-selectel-datasource-global-router-zone-group-v1"
description: |-
  Provides a list of zone groups in the Global Router service using public API v1.
---

# selectel\_global\_router\_zone\_group\_v1

Provides a list of zone groups in the Global Router service using public API v1. Zone group is a logical association of [zones](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/global_router_zone_v1). A global router can only connect networks of the same zone group. For more information about global routers, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/).

## Example Usage

```hcl
data "selectel_global_router_zone_v1" "zone_1" {
  name    = "ru-3"
  service = "vpc"
}

data "selectel_global_router_zone_group_v1" "zone_group_1" {
  name = data.selectel_global_router_zone_v1.zone_1.groups[0].name
}
```

## Argument Reference

* `name` - (Required) Zone group name, for example, `public_rf`. Retrieved from the [selectel_global_router_zone_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/global_router_zone_v1) data source of the corresponding zone, or via the [List zone groups](https://docs.selectel.ru/en/api/global-router/#tag/Zone-groups/operation/getZoneGroupsList) method in the Global Router API.

## Attributes Reference

* `id` - Unique identifier of the zone group.
* `name` - Zone group name.
* `description` - Optional description for the zone group.
* `created_at` - Time when the zone group was created.
* `updated_at` - Time when the zone group was updated.
