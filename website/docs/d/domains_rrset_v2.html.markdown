---
layout: "selectel"
page_title: "Selectel: selectel_domains_rrset_v2"
sidebar_current: "docs-selectel-datasource-domains-rrset-v2"
description: |-
  Provides a RRSet info in Selectel DNS Hosting using public API v2.
---

# selectel\_domains\_rrset_v2

Provides a RRSet info in DNS Hosting. For more information about RRSet in DNS Hosting, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/records/).

## Example Usage

```hcl
data "selectel_domains_rrset_v2" "rrset_1" {
  name = "example.com."
  type = "A"
  zone_id = "zone_id"
  ptoject_id = "project_id"
}
```

## Argument Reference

* `name` - (Required) RRSet name.

* `type` - (Required) RRSet type.

* `zone_id` - (Required) Zone ID.

* `project_id` - (Required) Selectel project ID.

## Attributes Reference

* `name` - RRSet name.

* `type` - RRSet type.

* `zone_id` - Zone ID.

* `project_id` - Selectel project ID.

* `ttl` - RRSet TTL.

* `comment` - Comment for RRSet.

* `managed_by` - RRSet owner.

* `records` - Set of records:
  
  * `content` - Value for record.

  * `disabled` - Shows if record available or not.
