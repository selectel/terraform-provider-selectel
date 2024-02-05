---
layout: "selectel"
page_title: "Selectel: selectel_domains_domain_v1"
sidebar_current: "docs-selectel-resource-domains-domain-v1"
description: |-
  Creates and manages a domain in Selectel DNS Hosting using public API v1.
---

# selectel\_domains\_domain\_v1

~> **Note:** This is legacy resource. Using the `selectel_domains_zone_v2` resource. They utilize different API and created zones live on different authoritative servers. Zone created in v2 API is entirely new zone, and not available via v1 api and vice versa.

Creates and manages a domain in DNS Hosting using public API v1. For more information about domains, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/zones/).

## Example usage

```hcl
resource "selectel_domains_domain_v1" "domain_1" {
  name = "example.com"
}
```

## Argument Reference

* `name` - (Required) Domain name. Changing this creates a new domain name.

## Attributes Reference

* `name` - Domain name.

* `user_id` - Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

## Import

You can import a domain:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_domains_domain_v1.domain_1 <domain_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<domain_id>` — Unique identifier of the domain, for example, `45623`. To get the domain ID, in the [Control panel](https://my.selectel.ru/network/domains/), go to **Networks Services** ⟶ **DNS Hosting** ⟶ the domain page ⟶ copy the domain ID from the address bar.