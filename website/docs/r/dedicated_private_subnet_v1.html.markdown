---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_private_subnet_v1"
sidebar_current: "docs-selectel-resource-dedicated-private-subnet-v1"
description: |-
  Creates and manages a private subnet in Selectel Dedicated Servers.
---

# selectel\_dedicated\_private\_subnet\_v1

Creates and manages a private subnet in Selectel Dedicated Servers.

## Example usage

```hcl
resource "selectel_dedicated_private_subnet_v1" "subnet_1" {
  location_id = "73bc417f-bc6b-45c1-8e06-ea9d5cce061c"
  vlan        = "100"
  subnet      = "192.168.100.0/24"
}
```

## Argument Reference

* `location_id` - (Required) Unique identifier of the location where the private subnet will be created. Retrieved from the [dedicated_location_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/dedicated_location_v1) data source.

* `vlan` - (Required) VLAN ID for the private subnet. Must be a unique VLAN within the location.

* `subnet` - (Required) CIDR block for the private subnet. Must be within private IP ranges: 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier of the private subnet.
* `vlan` - VLAN ID of the private subnet.
* `subnet` - CIDR block of the private subnet.

## Import

You can import a private subnet:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
terraform import selectel_dedicated_private_subnet_v1.subnet_1 <subnet_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/servers), go to **Servers and colocation** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<subnet_id>` — Unique identifier of the private subnet.
