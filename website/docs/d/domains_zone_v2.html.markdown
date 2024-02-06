---
layout: "selectel"
page_title: "Selectel: selectel_domains_zone_v2"
sidebar_current: "docs-selectel-datasource-domains-zone-v2"
description: |-
  Provides a zone info in Selectel DNS Hosting using public API v2.
---

# selectel\_domains\_zone_v2

Provides a zone info in DNS Hosting (API v2). For more information about zones in DNS Hosting, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/zones/).

## Example Usage

```hcl
data "selectel_domains_zone_v2" "zone_1" {
  name = "example.com."
}
```

With specific project id.

```hcl
data "selectel_domains_zone_v2" "zone_1" {
  name = "example.com."
  project_id = "project_id"
}
```

## Argument Reference

* `name` - (Required) Zone name.

* `project_id` - (Required) Selectel project ID.

## Attributes Reference
  
* `name` - Zone name.

* `project_id` - Selectel project id.

* `comment` - Comment for zone.

* `created_at` - Timestamp when zone was created.

* `updated_at` - Timestamp when zone was updated.

* `delegation_checked_at` - Timestamp of last delegation status check.

* `last_check_status` - Shows if zone delegated to selectel NS servers or not.

* `last_delegated_at` - Timestamp of last delegation status check when zone was delegated to selectel NS server.

* `disabled` - Shows if zone available or not.
