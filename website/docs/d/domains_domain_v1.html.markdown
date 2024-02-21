---
layout: "selectel"
page_title: "Selectel: selectel_domains_domain_v1"
sidebar_current: "docs-selectel-datasource-domains-domain-v1"
description: |-
  Provides an ID of a domain in Selectel DNS Hosting (legacy).
---

# selectel\_domains\_domain_v1

**WARNING**: This data source is applicable to DNS Hosting (legacy). We do not support and develop DNS Hosting (legacy), but domains and records created in DNS Hosting (legacy) continue to work until further notice. We recommend to transfer your data to DNS Hosting (actual). For more infomation about DNS Hosting (actual), see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/about-dns/).

Provides an ID of a domain in DNS Hosting (legacy). For more information about domains in DNS Hosting, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/zones/).

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
