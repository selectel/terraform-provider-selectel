---
layout: "selectel"
page_title: "Selectel: selectel_global_router_quota_v1"
sidebar_current: "docs-selectel-datasource-global-router-quota-v1"
description: |-
  Provides a list of quotas in the Global Router service using public API v1.
---

# selectel\_global\_router\_quota\_v1

Provides a list of quotas in the Global Router service using public API v1. For more information about service limits and restrictions, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/about-global-router/#limits).

## Example Usage

```hcl
data "selectel_global_router_quota_v1" "quota_1" {
  name        = "routers"
  scope       = "account_id"
  scope_value = "12345"
}
```

## Argument Reference

* `name` - (Required) Name of a resource under the quota. Available names are `routers`, `networks`, `subnets`, and `static_routes`.
* `scope` - (Optional) Quota scope. Global router quotas are currently applied only to an account level, the only available scope value is `account`.
* `scope_value` - (Optional) Unique identifier for the specified `scope`, for account level it's Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

## Attributes Reference

* `id` - Unique identifier of the quota.
* `name` - Name of the resource under the quota.
* `scope` - Scope of the quota. The only possible value now is `account`.
* `scope_value` - Scope value (Selectel account ID).
* `limit` - Quota limit, the maximum number of the specified resource that can be created.
