---
layout: "selectel"
page_title: "Selectel: selectel_global_router_zone_v1"
sidebar_current: "docs-selectel-datasource-global-router-zone-v1"
description: |-
  Provides a list of zones in the Global Router service using public API v1.
---

# selectel\_global\_router\_zone\_v1

Provides a list of zones in the Global Router service using public API v1. A zone represents a logical grouping of network resources that are used by a service or a product within one pool. Zones are logically aggregated into [zone groups](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/global_router_zone_group_v1). For example, cloud platform networks in the `ru-3` pool belong to the `ru-3` zone. For more information about global routers, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/).

## Example Usage

```hcl
data "selectel_global_router_zone_v1" "zone_1" {
  name    = "ru-3"
  service = "vpc"
}
```

## Argument Reference

* `name` - (Required) Pool, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `service` - (Optional) Name of the service.
                         Available names are `vpc`, `dedicated`, `vmware`, and `infra`.

## Attributes Reference

* `id` - Unique identifier of the zone.
* `name` - Name of the zone.
* `service` - Service name of the zone.
* `visible_name` - Name of the zone to display in the Control Panel.
* `enable` - Boolean flag, indicates whether networks in the zone can be created, updated, or deleted. `False` means that the zone is temporarily under maintenance and networks in it cannot be created, updated, or deleted.
* `allow_create` - Boolean flag, indicates whether the network can be created in the zone. `False` means that the zone is temporarily under maintenance and networks cannot be created in it.
* `allow_update` - Boolean flag, indicates whether the network can be updated in the zone. `False` means that the zone is temporarily under maintenance and networks in it cannot be updated.
* `allow_delete` - Boolean flag, indicates whether the network in the zone can be deleted. `False` means that the zone is temporarily under maintenance and networks in it cannot be deleted.
* `created_at` - Time when the zone was created.
* `updated_at` - Time when the zone was updated.
* `options` - Zone custom options.
* `groups` - List of zone groups that include this zone.
  * `id` - Unique identifier of the zone group.
  * `name` - Zone group name.
  * `description` - Optional description of the zone group.
  * `created_at` - Time when the zone group was created.
  * `updated_at` - Time when the zone group was updated.
