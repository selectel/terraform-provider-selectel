---
layout: "selectel"
page_title: "Selectel: selectel_domains_domain_v1"
sidebar_current: "docs-selectel-datasource-domains-domain-v1"
description: |-
  Provides an ID of a domain in Selectel DNS Hosting.
---

# selectel\_domains\_domain_v1

Provides an ID of a domain in DNS Hosting. For more information about domains in DNS Hosting, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/domains/).

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
