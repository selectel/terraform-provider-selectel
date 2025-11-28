---
layout: "selectel"
page_title: "Selectel: selectel_private_dns_zone_v1"
sidebar_current: "docs-selectel-private-dns_zone-v1"
description: |-
  Creates and manages a dns zone in Selectel Private DNS using public API v1.
---

# selectel\_private\_dns\_zone\_v1

Creates and manages a private dns zone using public API v1. For more information about private dns zones, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/private-dns/zone).

## Example usage

```hcl
resource "selectel_private_dns_zone_v1" "zone" {
    region = "ru-1"
    project_id = selectel_vpc_project_v2.project_1.id
	domain = "example.com."
	ttl = 3600
	records {
		domain = "sub.example.com."
		type = "A"
		ttl = 10
		values = [
			"192.168.0.1",
		]
	}
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the backup plan is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `domain` - (Required) Zone DNS domain, must be an FQDN.

* `ttl` - (Optional) Time to live in seconds for zone, default value is 3600.

* `records` - (Optional) List of zone records.
  * `domain` - (Required) Record dns domain, must be an FQDN.
  * `type` -  (Required) Record type. Available record types is `A`, `AAAA`, `MX`, `TXT`, `CNAME`.
  * `ttl` - (Optional) Time to live in seconds for record, if not specifed used zone ttl value.
  * `values` - (Required) List of values for a record.

## Attributes Reference

* `serial_number` - Zone SOA serial number.