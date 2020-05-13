---
layout: "selectel"
page_title: "Selectel: selectel_domains_domain_v1"
sidebar_current: "docs-selectel-resource-domains-domain-v1"
description: |-
  Manages a V1 domain resource within Selectel Domains API Service.
---

# selectel\_domains\_domain\_v1

Manages a V1 domain resource within Selectel Domains API Service.

## Example usage

```hcl
resource "selectel_domains_domain_v1" "domain_1" {
  name = "test-domain.xyz"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the domain.
  Changing this creates a new domain name.

## Attributes Reference

The following attributes are exported:

* `name` - The name of the domain.

* `user_id` - Identifier of the Selectel API user.

## Import

Domain can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selectel_domains_domain_v1.domain_1 45623
```
