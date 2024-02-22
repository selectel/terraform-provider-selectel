---
layout: "selectel"
page_title: "Selectel: selectel_domains_rrset_v2"
sidebar_current: "docs-selectel-resource-domains-rrset-v2"
description: |-
  Creates and manages an RRSet in Selectel DNS Hosting (actual) using public API v2.
---

# selectel\_domains\_rrset\_v2

Creates and manages an RRSet in DNS Hosting (actual) using public API v2. For more information about RRSets, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/records/).

## Example usage

### A RRSet

```hcl
resource "selectel_domains_rrset_v2" "a_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "A"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "127.0.0.1"
    # The content value is "<ipv4_address>"
  }
}
```

### AAAA RRSet

```hcl
resource "selectel_domains_rrset_v2" "aaaa_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "AAAA"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "2400:cb00:2049:1::a29f:1804"
    # The content value is "<ipv6_address>"
  }
}
```

### TXT RRSet

```hcl
resource "selectel_domains_rrset_v2" "txt_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "TXT"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "\"hello, world!\""
    # The content value is "<text>"
  }
}
```

### CNAME RRSet

```hcl
resource "selectel_domains_rrset_v2" "cname_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "CNAME"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "origin.com."
    # The content value is "<target>"
  }
}
```

### MX RRSet

```hcl
resource "selectel_domains_rrset_v2" "mx_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "MX"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "10 mail.example.org."
    # The content value is "<priority> <host>"
  }
}
```

### SRV RRSet

```hcl
resource "selectel_domains_rrset_v2" "srv_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "_sip._tcp.example.com."
  type       = "SRV"
  ttl        = 120
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "10 20 30 example.org."
    # The content value is "<priority> <weight> <port> <target>"
  }
}
```

### SSHFP RRSet

```hcl
resource "selectel_domains_rrset_v2" "sshfp_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "SSHFP"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "1 1 7491973e5f8b39d5327cd4e08bc81b05f7710b49"
    # The content value is "<algorithm> <fingerprint_type> <fingerprint>"
  }
}
```

### ALIAS RRSet

```hcl
resource "selectel_domains_rrset_v2" "alias_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "ALIAS"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "origin.com."
    # The content value is "<target>"
  }
}
```

### CAA RRSet

```hcl
resource "selectel_domains_rrset_v2" "caa_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "CAA"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "128 issue \"letsencrypt.com.\""
    # The content value is "<flag> <tag> <value>"
  }
}
```

## Argument Reference

* `zone_id` - (Required) Unique identifier of the zone. Changing this creates a new RRSet. Retrieved from the [selectel_domains_zone_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/domains_zone_v2) resource.

* `name` - (Required) RRSet name. Changing this creates a new RRSet. The value must be the same as the zone name. If `type` is `SRV`, the name must also include service and protocol, see the [example usage for SRV RRSet](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/domains_rrset_v2#srv-rrset).

* `type` - (Required) RRSet type. Changing this creates a new RRSet. Available types are `A`, `AAAA`, `TXT`, `CNAME`, `MX`, `SRV`, `SSHFP`, `ALIAS`, `CAA`.

* `ttl` - (Required) RRSet time-to-live in seconds. The available range is from 60 to 604800.

* `records` - (Required) List of records in the RRSet.
  
  * `content` - (Required) Record value. The value depends on the RRSet type.
    - `<ipv4_address>` — IPv4-address. Applicable only to A RRSets.
    - `<ipv6_address>` — IPv6-address. Applicable only to AAAA RRSets.
    - `<text>` — Any text wrapped in `\"`. Applicable only to TXT RRSets.
    - `<target>` — Canonical name of the host providing the service with a dot at the end. Applicable only to CNAME, ALIAS, and SRV RRSets.
    - `<name_server>` — Canonical name of the NS server. Applicable only to NS RRSets.
    - `<priority>` — Priority of the records preferences. Applicable only to MX and SRV RRSets. Lower value means more preferred.
    - `<host>` — Name of the mailserver with a dot at the end. Applicable only to MX RRSets.
    - `<weight>` — Weight for the records with the same priority. Higher value means more preferred. Applicable only to SRV RRSets.
    - `<port>` — TCP or UDP port of the host of the service. Applicable only to SRV RRSets.
    - `<algorithm>` — Algorithm of the public key. Applicable only to SSHFP RRSets. Available values are `1` for RSA, `2` for DSA, `3` for ECDSA, `4` for Ed25519.
    - `<fingerprint_type>` — Algorithm used to hash the public key. Applicable only to SSHFP RRSets. Available values are `1` for SHA-1, `2` for SHA-256.
    - `<fingerprint>` — Hexadecimal representation of the hash result, as text. Applicable only to SSHFP RRSets.
    - `<flag>` — Critical value that has a specific meaning per RFC. Applicable only to CAA RRSets. The available range is from 0 to 128.
    - `<tag>` — Identifier of the property represented by the record. Applicable only to CAA RRSets. Available values are `issue`, `issuewild`, `iodef`, `auth`, `path`, `policy`.
    - `<value>` — Value associated with the tag wrapped in `\"`. Applicable only to CAA RRSets.

  * `disabled` - (Optional) Enables or disables the record. Boolean flag, the default value is false.

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new RRSet. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `comment` - (Optional) Comment to add to the RRSet.

## Attributes Reference

* `managed_by` - RRSet owner.

## Import

You can import an RRSet:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export SEL_PROJECT_ID=<selectel_project_id>
terraform import selectel_domains_rrset_v2.rrset_1 <zone_name>/<rrset_name>/<rrset_type>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to Cloud Platform ⟶ project name ⟶ copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/craas/about/projects/).

* `<zone_name>` — Zone name, for example, `example.com.`. To get the name, in the [Control panel](https://my.selectel.ru/dns/), go to **DNS**. The zone name is in the **Zone** column.

* `<rrset_name>` — RRSet name, for example, `example.com.`. To get the name, in the [Control panel](https://my.selectel.ru/dns/), go to **DNS** → the zone page. The RRSet name is in the **Group name** column.

* `<rrset_type>` — RRSet type. To get the type, in the [Control panel](https://my.selectel.ru/dns/), go to **DNS** → the zone page. The RRSet type is in the **Type** column.
