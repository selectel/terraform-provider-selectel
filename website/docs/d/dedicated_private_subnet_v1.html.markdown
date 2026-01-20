---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_private_subnet_v1"
sidebar_current: "docs-selectel-data-source-dedicated-private-subnet-v1"
description: |-
  Retrieves information about dedicated private subnets.
---

# selectel\_dedicated\_private\_subnet\_v1

Retrieves information about dedicated private subnets.

## Example usage

```hcl
data "selectel_dedicated_private_subnet_v1" "subnet_ds" {
  project_id = selectel_vpc_project_v2.project_1.id

  filter {
    location_id = "73bc417f-bc6b-45c1-8e06-ea9d5cce061c"
    vlan        = "100"
    subnet      = "192.168.100.0/24"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource.

* `filter` - (Optional) Filter for searching subnets.
  * `location_id` - (Optional) Location ID to filter subnets by location.
  * `subnet` - (Optional) Subnet CIDR to filter (e.g., "192.168.100.0/24").
  * `vlan` - (Optional) VLAN ID to filter subnets by VLAN tag.
  * `ip` - (Optional) IP address to check if it's included in the subnet.

## Attributes Reference

* `id` - Unique identifier of the data source (checksum of subnet IDs).

* `subnets` - List of matching subnets, each containing:
  * `id` - Unique identifier of the subnet (UUID).
  * `subnet` - Subnet CIDR (e.g., "192.168.100.0/24").
  * `vlan` - VLAN ID (tag in API).
  * `reserved_ip` - List of reserved IP addresses in the subnet.

## Import

You can import a subnet data source:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
terraform import selectel_dedicated_private_subnet_v1.subnet_ds <subnet_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project.

* `<subnet_id>` — Unique identifier of the subnet.
