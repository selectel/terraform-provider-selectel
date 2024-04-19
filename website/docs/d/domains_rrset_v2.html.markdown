---
layout: "selectel"
page_title: "Selectel: selectel_domains_rrset_v2"
sidebar_current: "docs-selectel-datasource-domains-rrset-v2"
description: |-
  Provides information about an RRSet in Selectel DNS Hosting (actual).
---

# selectel\_domains\_rrset_v2

Provides information about an RRSet in DNS Hosting (actual). For more information about RRSets, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/records/).

## Example Usage

```hcl
data "selectel_domains_rrset_v2" "rrset_1" {
  name       = "example.com."
  type       = "A"
  zone_id    = selectel_domains_zone_v2.zone_1.id
  project_id = selectel_vpc_project_v2.project_1.id
}
```

## Argument Reference

* `name` - (Required) RRSet name.

* `type` - (Required) RRSet type. Available types are `A`, `AAAA`, `TXT`, `CNAME`, `NS`, `MX`, `SRV`, `SSHFP`, `ALIAS`, `CAA`.

* `zone_id` - (Required) Unique identifier of the zone. Retrieved from the [selectel_domains_zone_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/domains_zone_v2) resource.

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

## Attributes Reference

* `ttl` - RRSet time-to-live in seconds.

* `comment` - Comment for the RRSet.

* `managed_by` - RRSet owner.

* `records` - List of records in the RRSet.
  
  * `content` - Record value.

  * `disabled` - Shows if the record is enabled or disabled.
