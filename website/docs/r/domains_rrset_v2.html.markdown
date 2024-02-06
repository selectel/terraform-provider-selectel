---
layout: "selectel"
page_title: "Selectel: selectel_domains_rrset_v2"
sidebar_current: "docs-selectel-resource-domains-rrset-v2"
description: |-
  Creates and manages a RRSet in Selectel DNS Hosting using public API v2.
---

# selectel\_domains\_rrset\_v2

Creates and manages a RRSet in DNS Hosting using public API v2. For more information about RRSet, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/records/).

## Example usage

### A RRSet

```hcl
resource "selectel_domains_rrset_v2" "a_rrset_1" {
  zone_id = "zone_id"
  name    = "a.example.com."
  type    = "A"
  ttl     = 60
  records {
    content  = "127.0.0.1"
  }
  records {
    content  = "127.0.0.2"
    disabled = true
  }
}

```

### AAAA RRSet

```hcl
resource "selectel_domains_rrset_v2" "aaaa_rrset_1" {
  zone_id   = "zone_id"
  name      = "aaaa.example.com."
  type      = "AAAA"
  ttl       = 60
  records {
    content = "2400:cb00:2049:1::a29f:1804"
  }  
}

```

### TXT RRSet

```hcl
resource "selectel_domains_rrset_v2" "txt_rrset_1" {
  zone_id   = "zone_id"
  name      = "txt.example.com."
  type      = "TXT"
  ttl       = 60
  records {
    content   = "\"hello, world!\""
  } 
}
```

### CNAME RRSet

```hcl
resource "selectel_domains_rrset_v2" "cname_rrset_1" {
  zone_id   = "zone_id"
  name      = "cname.example.com."
  type      = "CNAME"
  ttl       = 60
  records {
    content = "origin.com."
  }
}
```

### NS RRSet

```hcl
resource "selectel_domains_rrset_v2" "ns_rrset_1" {
  zone_id   = "zone_id"
  name      = "example.com."
  type      = "NS"
  ttl       = 86400
  records {
    content = "ns5.selectel.org"
  }
}
```

### MX RRSet

Content includes: "priority host"

```hcl
resource "selectel_domains_rrset_v2" "mx_rrset_1" {
  zone_id   = "zone_id"
  name      = "mx.example.com."
  type      = "MX"
  ttl       = 60
  records {
    content = "10 mail.example.org."
  }
}
```

### SRV RRSet

Content includes: "priority weight port target"

```hcl
resource "selectel_domains_rrset_v2" "srv_rrset_1" {
  zone_id   = "zone_id"
  name      = "_sip._tcp.example.com."
  type      = "SRV"
  ttl       = 120
  records {
    content = "10 20 30 mail.example.org."
  }
}
```

### SSHFP RRSet

Content includes: "algorithm fingerprint_type fingerprint"

```hcl
resource "selectel_domains_rrset_v2" "sshfp_rrset_1" {
  zone_id    = "zone_id"
  name       = "sshfp.example.com."
  type       = "SSHFP"
  ttl         = 60
  records {
    content  = "1 1 7491973e5f8b39d5327cd4e08bc81b05f7710b49"
  }
}
```

### ALIAS RRSet

```hcl
resource "selectel_domains_rrset_v2" "alias_rrset_1" {
  zone_id   = "zone_id"
  name      = "example1.com."
  type      = "ALIAS"
  ttl       = 60
  records {
    content = "example2.com."
  }
}
```

### CAA RRSet

Content includes: "flag tag value"

```hcl
resource "selectel_domains_rrset_v2" "caa_rrset_1" {
  zone_id   = "zone_id"
  name      = "example.com."
  type      = "CAA"
  ttl       = 60
  records {
    content = "128 issue \"letsencrypt.com.\""
  }
}
```

## Argument Reference

* `zone_id` - (Required) Zone ID.

* `name` - (Required) Name of the zone RRSet. The name format depends on the RRSet type, see the examples above.

* `type` - (Required) Type of the RRSet.

* `ttl` - (Required) Time-to-live for the RRSet in seconds. The available range is from 60 to 604800.

* `records` - (Required) Set of records:
  
  * `content` - (Required) Value for record.

  * `disabled` - (Optional, default: false) Shows if record available or not.

* `project_id` - (Required) Selectel project ID.

* `comment` - (Optional) Comment for RRSet.

## Attributes Reference

* `zone_id` - Zone ID.

* `name` - Name of the RRSet.

* `type` - Type of the RRSet.

* `ttl` - Time-to-live for the RRSet in seconds.

* `records` - Set of records:
  
  * `content` - Value for record.

  * `disabled` - Shows if record available or not.

* `project_id` - Selectel project ID.

* `comment` - Comment for RRSet.

* `managed_by` - RRSet owner.

## Import

You can import a RRSet:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export SEL_PROJECT_ID=<project_id>
terraform import selectel_domains_rrset_v2.rrset_1 <zone_name>/<rrset_name>/<rrset_type>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Selectel project ID.

* `<zone_name>` — Zone name.

* `<rrset_name>` — RRSet name.

* `<rrset_type>` — Type of RRSet.
