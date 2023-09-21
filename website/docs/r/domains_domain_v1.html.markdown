---
layout: "selectel"
page_title: "Selectel: selectel_domains_domain_v1"
sidebar_current: "docs-selectel-resource-domains-domain-v1"
description: |-
  Creates and manages a domain in Selectel DNS Hosting using public API v1.
---

# selectel\_domains\_domain\_v1

Creates and manages a domain in DNS Hosting using public API v1. For more information about domains, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/domains/).

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
$ export OS_DOMAIN_NAME=999999
$ export OS_USERNAME=example_user
$ export OS_PASSWORD=example_password
$ terraform import selectel_domains_domain_v1.domain_1 <domain_id>
```

where `<domain_id>` is a unique identifier of the domain, for example, `45623`. To get the domain ID, in the [Control panel](https://my.selectel.ru/network/domains/), go to **Networks Services** ⟶ **DNS Hosting** ⟶ the domain page ⟶ copy the domain ID from the address bar.

### Environment Variables

For import, you must set the environment variable `SEL_TOKEN=<selectel_api_token>`,

where `<selectel_api_token>` is a Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).
