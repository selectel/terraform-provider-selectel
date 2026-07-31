---
layout: "selectel"
page_title: "Selectel: selectel_networking_subnetpool_v2"
sidebar_current: "docs-selectel-resource-networking-subnetpool-v2"
description: |-
  Creates and manages a subnet pool in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_subnetpool\_v2

Creates and manages a subnet pool using Neutron API v2.0. Subnets allocate their CIDR from a pool instead of specifying it explicitly — see `subnetpool_id` and `prefix_length` of the [selectel_networking_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/networking_subnet_v2) resource.

~> **Note:** `shared` and `is_default` are writable only by the Keystone admin role, which no granular Selectel VPC role grants, so this resource reports them as read-only attributes. `ip_version` is derived by Neutron from `prefixes` and is not accepted on create either.

## Example Usage

```hcl
resource "selectel_networking_subnetpool_v2" "subnetpool_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-3"
  name              = "subnetpool-1"
  prefixes          = ["192.168.200.0/24"]
  default_prefixlen = 26
  min_prefixlen     = 24
  max_prefixlen     = 28
  tags              = ["production"]
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new subnet pool.

* `region` - (Required) Pool where the subnet pool is located, for example, `ru-3`. Changing this creates a new subnet pool.

* `name` - (Required) Name of the subnet pool. Can be changed without replacement.

* `prefixes` - (Required) Non-empty list of CIDRs the pool allocates from. The list can be extended without replacement.

* `description` - (Optional) Description of the subnet pool.

* `default_quota` - (Optional) Maximum number of addresses that can be allocated from the pool. If you do not specify the value, the API assigns one.

* `default_prefixlen` - (Optional) Prefix length used when a subnet requests an allocation without specifying `prefix_length`. If you do not specify the value, the API assigns one.

* `min_prefixlen` - (Optional) Smallest prefix length the pool allows. If you do not specify the value, the API assigns one.

* `max_prefixlen` - (Optional) Largest prefix length the pool allows. If you do not specify the value, the API assigns one.

* `address_scope_id` - (Optional) Unique identifier of the address scope the pool belongs to. Clearing the value detaches the pool from its address scope.

* `tags` - (Optional) Tags of the subnet pool that Terraform manages. Tags added outside Terraform are reported in `all_tags` and are not removed by this resource.

## Attributes Reference

* `ip_version` - Internet protocol version of the pool, derived by Neutron from `prefixes`.

* `all_tags` - All tags of the subnet pool, including the ones added outside Terraform.

* `shared` - Specifies whether the pool can be used by any project. Read-only.

* `is_default` - Specifies whether the pool is the default one for its address family. Read-only.

* `created_at` - Time when the subnet pool was created.

* `updated_at` - Time when the subnet pool was last updated.

* `revision_number` - Revision number of the subnet pool.

## Behavior notes

A pool that existing subnets still allocate from cannot be deleted: Neutron refuses with a conflict, the delete reports a diagnostic error, and the resource stays in the state. Delete the dependent subnets first.

## Import

You can import a subnet pool:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_networking_subnetpool_v2.subnetpool_1 <subnetpool_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<pool>` — Pool where the subnet pool is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<subnetpool_id>` — Unique identifier of the subnet pool, for example, `3c9f1a7e-4b2d-4e60-8a1b-2c3d4e5f6a7b`.
