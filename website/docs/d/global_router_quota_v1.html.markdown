---
layout: "selectel"
page_title: "Selectel: selectel_global_router_quota_v1"
sidebar_current: "docs-selectel-datasource-global-router-quota-v1"
description: |-
  Provides a list of quotas in the Selectel Global Router service using public API v1.
---

# selectel\_global\_router\_quota\_v1

Provides a list of quotas in the Selectel Global Router service using public API v1. For more information about quotas, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/about-global-router/#limits).

## Example Usage

```hcl
data "selectel_global_router_quota_v1" "quota_1" {
  name        = "routers"
  scope       = "account_id"
  scope_value = "12345"
}
```

## Argument Reference

* `name` - (Required) The resource name for which to display the quota value. Available names are: `routers`, `networks`, `subnets`, and `static_routes`.
* `scope` - (Optional) Scope of the quota. Global Router quotas are currently applied on account level, so scope value is `account_id`.
* `scope_value` - (Optional) Unique identifier for the specified `scope`, for account level it's account ID, for example `12345`.

## Attributes Reference

* `id` - Unique identifier of the quota.
* `name` - Name of the resource under the quota.
* `scope` - Scope of the quota (the only possible value now is `account`).
* `scope_value` - Scope value (account ID).
* `limit` - Quota limit, the maximum number of the specified resource available for creation.
