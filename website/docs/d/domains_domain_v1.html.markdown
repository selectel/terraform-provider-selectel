---
layout: "selectel"
page_title: "Selectel: selectel_domains_domain_v1"
sidebar_current: "docs-selectel-datasource-domains-domain-v1"
description: |-
  Provides an ID of a domain in Selectel DNS Hosting (legacy).
---

# selectel\_domains\_domain_v1

!> **WARNING:** This data source is deprecated and will be removed in a future major version. Use selectel_domains_zone_v2 instead.

DNS Hosting (legacy) is not supported or developed anymore. For more information about DNS Hosting (actual), see the [official Selectel documentation](https://docs.selectel.ru/en/networks-services/dns/about-dns/).

Provides an ID of a domain in DNS Hosting (legacy).

## Example Usage

```hcl
data "selectel_domains_domain_v1" "domain_1" {
  name = "example.com"
}
```

## Argument Reference

* `name` - (Required) Domain name.

## Attributes Reference

* `id` - Unique identifier of the domain.

* `name` - Domain name.

* `user_id` - Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).
