---
layout: "selectel"
page_title: "Selectel: selectel_domains_zone_v2"
sidebar_current: "docs-selectel-resource-domains-zone-v2"
description: |-
  Creates and manages a zone in Selectel DNS Hosting using public API v2.
---

# selectel\_domains\_zone\_v2

Creates and manages a zone in DNS Hosting using public API v2. For more information about zones, see the [official Selectel documentation](https://docs.selectel.ru/networks-services/dns/zones/).

## Example usage

```hcl
resource "selectel_domains_zone_v2" "zone_1" {
  name = "example.com."
  project_id = "project_id"
}
```

## Argument Reference

* `name` - (Required) Zone name. Changing this creates a new zone name.

* `project_id` - (Required) Selectel project id. Scope for creating zone.

* `comment` - (Optional) Comment for zone.

* `disabled` - (Optional) Set zone available or not.

## Attributes Reference

* `id` - Unique identifier of the zone.

* `name` - Zone name.

* `project_id` - Selectel project id.

* `comment` - Comment for zone.

* `created_at` - Timestamp when zone was created.

* `updated_at` - Timestamp when zone was updated.

* `delegation_checked_at` - Timestamp of last delegation status check.

* `last_check_status` - Shows if zone delegated to selectel NS servers or not.

* `last_delegated_at` - Timestamp of last delegation status check when zone was delegated to selectel NS server.

* `disabled` - Shows if zone available or not.

## Import

You can import a zone:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export SEL_PROJECT_ID=<project_id>
terraform import selectel_domains_zone_v2.zone_1 <zone_name>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Selectel project ID.

* `<zone_name>` — Zone name.
