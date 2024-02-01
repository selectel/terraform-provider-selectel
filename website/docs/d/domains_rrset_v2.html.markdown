---
layout: "selectel"
page_title: "Selectel: selectel_domains_rrset_v2"
sidebar_current: "docs-selectel-datasource-domains-rrset-v2"
description: |-
  Provides a rrset info in Selectel DNS Hosting using public API v2.
---

# selectel\_domains\_rrset_v2

Provides a rrset info in DNS Hosting. For more information about rrset in DNS Hosting, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/records/).

## Example Usage

```hcl
data "selectel_domains_rrset_v2" "rrset_1" {
  name = "example.com."
  type = "A"
  zone_id = "zone_id"
}
```

With specific project id.

```hcl
data "selectel_domains_rrset_v2" "rrset_1" {
  name = "example.com."
  type = "A"
  zone_id = "zone_id"
  ptoject_id = "project_id"
}
```

## Argument Reference

* `name` - (Required) Rrset name.

* `type` - (Required) Rrset type.

* `zone_id` - (Required) Zone ID.

* `project_id` - (Optional) Selectel project ID.

## Attributes Reference

* `name` - Rrset name.

* `type` - Rrset type.

* `zone_id` - Zone ID.

* `project_id` - Selectel project ID.

* `ttl` - Rrset TTL.

* `comment` - Comment for rrset.

* `managed_by` - Rrset owner.

* `records` - Set of records:
  
  * `content` - Value for record.

  * `disabled` - Shows if record available or not.
