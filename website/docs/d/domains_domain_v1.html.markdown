---
layout: "selectel"
page_title: "Selectel: selectel_domains_domain_v1"
sidebar_current: "docs-selectel-datasource-domains-domain-v1"
description: |-
  Get information on Selectel Domains domain.
---

# selectel\_domains\_domain_v1

Use this data source to get the ID of an available domain object within Selectel Domains API Service.

## Example Usage

```hcl
data "selectel_domains_domain_v1" "domain_1" {
  name = "test-domain.xyz"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the domain.

## Attributes Reference

`id` is set to the ID of the found domain. In addition, the following attributes
are exported:

* `name` - The name of the domain.

* `user_id` - Identifier of the Selectel API user.
