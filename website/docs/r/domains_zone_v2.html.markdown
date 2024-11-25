---
layout: "selectel"
page_title: "Selectel: selectel_domains_zone_v2"
sidebar_current: "docs-selectel-resource-domains-zone-v2"
description: |-
  Creates and manages a zone in Selectel DNS Hosting (actual) using public API v2.
---

# selectel\_domains\_zone\_v2

Creates and manages a zone in DNS Hosting (actual) using public API v2. For more information about zones, see the [official Selectel documentation](https://docs.selectel.ru/en/networks-services/dns/zones/).

## Example usage

```hcl
resource "selectel_domains_zone_v2" "zone_1" {
  name       = "example.com."
  project_id = selectel_vpc_project_v2.project_1.id
}
```

## Argument Reference

* `name` - (Required) Zone name. Changing this creates a new zone.

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `comment` - (Optional) Comment to add to the zone.

* `disabled` - (Optional) Enables or disables the zone. Boolean flag, the default value is false.

## Attributes Reference

* `created_at` - Time when the zone was created in the RFC 3339 timestamp format.

* `updated_at` - Time when the zone was updated in the RFC 3339 timestamp format.

* `delegation_checked_at` - Time when DNS Hosting checked if the zone was delegated to Selectel NS servers in the RFC 3339 timestamp format.

* `last_check_status` - Zone status retrieved during the last delegation check.

* `last_delegated_at` - Equals to the `delegation_check_at` argument value when the `last_check_status` is `true`.

## Import

You can import a zone:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
terraform import selectel_domains_zone_v2.zone_1 <zone_name>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/dns), go to **DNS** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<zone_name>` — Zone name, for example, `example.com.`. To get the name, in the [Control panel](https://my.selectel.ru/dns/), go to **DNS**. The zone name is in the **Zone** column.
