---
layout: "selectel"
page_title: "Selectel: selectel_domains_zone_v2"
sidebar_current: "docs-selectel-datasource-domains-zone-v2"
description: |-
  Provides information about a zone in Selectel DNS Hosting (actual).
---

# selectel\_domains\_zone_v2

Provides information about a zone in Selectel DNS Hosting (actual). For more information about zones, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/zones/).

## Example Usage

```hcl
data "selectel_domains_zone_v2" "zone_1" {
  name       = "example.com."
  project_id = selectel_vpc_project_v2.project_1.id
}
```

## Argument Reference

* `name` - (Required) Zone name.

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

## Attributes Reference

* `comment` - Comment for the zone.

* `created_at` - Time when the zone was created in the RFC 3339 timestamp format.

* `updated_at` - Time when the zone was updated in the RFC 3339 timestamp format.

* `delegation_checked_at` - Time when DNS Hosting checked if the zone was delegated to Selectel NS servers in the RFC 3339 timestamp format.

* `last_check_status` - Zone status retrieved during the last delegation check.

* `last_delegated_at` - Equals to the `delegation_check_at` argument value when the `last_check_status` is `true`.

* `disabled` - Shows if the zone is enabled or disabled.
