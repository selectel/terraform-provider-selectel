---
layout: "selectel"
page_title: "Selectel: selectel_domains_record_v1"
sidebar_current: "docs-selectel-resource-domains-record-v1"
description: |-
  Creates and manages a record in Selectel DNS Hosting using public API v1.
---

# selectel\_domains\_record\_v1

**WARNING**: This resource is applicable to DNS Hosting (legacy). We do not support and develop DNS Hosting (legacy), but domains and records created in DNS Hosting (legacy) continue to work until further notice. We recommend to transfer your data to DNS Hosting (actual). For more infomation about DNS Hosting (actual), see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/about-dns/).
To create records in DNS Hosting (actual) use the [selectel_domains_rrset_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/selectel_domains_rrset_v2) resource.

Creates and manages a record in DNS Hosting (legacy) using public API v1. For more information about records, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/records/).

## Example usage

### A Record

```hcl
resource "selectel_domains_record_v1" "a_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "A"
  content   = "127.0.0.1"
  ttl       = 60
}
```

### AAAA Record

```hcl
resource "selectel_domains_record_v1" "aaaa_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "AAAA"
  content   = "2400:cb00:2049:1::a29f:1804"
  ttl       = 60
}
```

### TXT Record

```hcl
resource "selectel_domains_record_v1" "txt_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "TXT"
  content   = "hello, world!"
  ttl       = 60
}
```

### CNAME Record

```hcl
resource "selectel_domains_record_v1" "cname_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "CNAME"
  content   = "origin.com"
  ttl       = 60
}
```

### NS Record

```hcl
resource "selectel_domains_record_v1" "ns_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "NS"
  content   = "ns5.selectel.org"
  ttl       = 86400
}
```

### MX Record

```hcl
resource "selectel_domains_record_v1" "mx_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "MX"
  content   = "mail.example.org"
  ttl       = 60
  priority  = 10
}
```

### SRV Record

```hcl
resource "selectel_domains_record_v1" "srv_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "SRV"
  ttl       = 120
  priority  = 10
  weight    = 20
  target    = "example.com"
  port      = 100
}
```

### SSHFP Record

```hcl
resource "selectel_domains_record_v1" "sshfp_record_1" {
  domain_id        = selectel_domains_domain_v1.main_domain.id
  name             = format("%s", selectel_domains_domain_v1.main_domain.name)
  type             = "SSHFP"
  ttl              = 60
  algorithm        = 1
  fingerprint_type = 1
  fingerprint      = "01AA"
}
```

### ALIAS Record

```hcl
resource "selectel_domains_record_v1" "alias_record_1" {
  domain_id = selectel_domains_domain_v1.main_domain.id
  name      = format("subc.%s", selectel_domains_domain_v1.main_domain.name)
  type      = "ALIAS"
  content   = format("%s", selectel_domains_domain_v1.main_domain.name)
  ttl       = 60
}
```

### CAA Record

```hcl
resource "selectel_domains_record_v1" "caa_record_1" {
  domain_id = selectel_domains_domain_v1.main_domain.id
  name      = format("caa.%s", selectel_domains_domain_v1.main_domain.name)
  type      = "CAA"
  ttl       = 60
  tag       = "issue"
  flag      = 128
  value     = "letsencrypt.com"
}
```

## Argument Reference

* `domain_id` - (Required) Unique identifier of the associated domain. Changing this creates a new domain record. Retrieved from the [selectel_domains_domain_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/domains_domain_v1) resource.

* `name` - (Required) Name of the domain record. The name format depends on the record type, see the examples above.

* `type` - (Required) Type of the record. Available values are `A`, `AAAA`, `TXT`, `CNAME`, `NS`, `MX`, `SRV`, `SSHFP`, `ALIAS`, `CAA`.

* `content` - (Optional) Content of the record. Not applicable to SRV, SSHFP, CAA records.

* `ttl` - (Required) Time-to-live for the record in seconds. The available range is from 60 to 604800.

* `priority` - (Optional) Priority of the records preferences. Applicable only to MX and SRV records. Lower value means more preferred.

* `weight` - (Optional) Weight for the records with the same priority. Higher value means more preferred. Applicable only to SRV records.

* `target` - (Optional) Canonical name of the host providing the service. Applicable only to SRV records.

* `port` - (Optional) TCP or UDP port of the host of the service. Applicable only to SRV records.

* `algorithm` - (Optional) Algorithm of the public key. Applicable only to SSHFP records. Available values are `RSA`, `DSA`, `ECDSA`, `Ed25519`.

* `fingerprint_type` - (Optional) Algorithm used to hash the public key. Applicable only to SSHFP records. Available values are `SHA-1`, `SHA-256`.

* `fingerprint` - (Optional) Hexadecimal representation of the hash result, as text. Applicable only to SSHFP records.

* `tag` - (Optional) Identifier of the property represented by the record. Applicable only to CAA records. Available values are `issue`, `issuewild`, `iodef`, `auth`, `path`, `policy`.

* `flag` - (Optional) Critical value that has a specific meaning per RFC. Applicable only to CAA records. The available range is from 0 to 128.

* `value` - (Optional) Value associated with the tag. Applicable only to CAA records.

* `email` - (Optional) Email of the domain administrator. Applicable only to SOA records.

## Attributes Reference

* `content` - Content of the record. Applicable only to A, AAAA, TXT, CNAME, NS, MX, ALIAS records.

* `priority` - Priority of the records preferences. Applicable only to MX, SRV records.

* `weight` - Weight for the records with the same priority. Applicable only to SRV records.

* `target` - Canonical name of the host providing the service. Applicable only to SRV records.

* `port` - TCP or UDP port of the host of the service. Applicable only to SRV records.

* `algorithm` - Algorithm of the public key. Applicable only to SSHFP records.

* `fingerprint_type` - Algorithm used to hash the public key. Applicable only to SSHFP records.

* `fingerprint` - Hexadecimal representation of the hash result, as text. Applicable only to SSHFP records.

* `tag` - Identifier of the property represented by the record. Applicable only to CAA records.

* `flag` - Critical value that has a specific meaning per RFC. Applicable only to CAA records.

* `value` - Value associated with the tag. Applicable only to CAA records.

* `email` - Email of the domain administrator. Applicable only to SOA records.

## Import

You can import a domain record:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_domains_record_v1.record_1 <domain_id>/<record_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<domain_id>` — Unique identifier of the domain, for example, `45623`. To get the domain ID, in the [Control panel](https://my.selectel.ru/network/domains/), go to **Networks Services** ⟶ **DNS Hosting** ⟶ the domain page ⟶ copy the domain ID from the address bar.

* `<record_id>` — Unique identifier of the record, for example, `123`. To get the record ID, use [DNS Hosting API](https://developers.selectel.ru/docs/cloud-services/dns_api/).