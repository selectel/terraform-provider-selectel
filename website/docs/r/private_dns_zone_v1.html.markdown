---
layout: "selectel"
page_title: "Selectel: selectel_private_dns_zone_v1"
sidebar_current: "docs-selectel-private-dns-zone-v1"
description: |-
  Creates and manages a DNS zone and record sets in Selectel Private DNS using public API v1
---

# selectel\_private\_dns\_zone\_v1

Creates and manages a private DNS zone and its record sets using public API v1. For more information about private DNS, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/private-dns/).

Within a single pool in a project, you can create no more than 100 zones. The maximum number of resource records in a zone is 1,000. For more information about Selectel infrastructure, see the [official Selectel documentation](https://docs.selectel.ru/en/infrastructure/locations/#pool). For more information about projects, see the [official Selectel documentation](https://docs.selectel.ru/en/access-control/projects/about-projects/).

## Example usage

```hcl
resource "selectel_private_dns_zone_v1" "zone_1" {
	region     = "ru-1"
	project_id = selectel_vpc_project_v2.project_1.id
	domain     = "example.com."
	records {
		domain = "sub.example.com."
		type   = "A"
		values = [
			"192.168.0.2",
		]
	}
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the DNS zone is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#private-dns).

* `domain` - (Required) Zone domain name, must be a fully qualified domain name (FQDN).

* `ttl` - (Optional) Time to live (TTL) in seconds for the zone, the default value is 3 600.

* `records` - (Optional) List of the zone record sets:
    * `domain` - (Required) Domain of the record set, must be an FQDN.
    * `type` -  (Required) Record set type. Available types are `A`, `AAAA`, `MX`, `TXT`, `CNAME`.
    * `ttl` - (Optional) Time to live (TTL) in seconds for the record. If not specifed, zone TTL is used for the record.
    * `values` - (Required) List of record set values.


## Attributes Reference

* `serial_number` - Zone SOA serial number.