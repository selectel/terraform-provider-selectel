---
layout: "selectel"
page_title: "Selectel: selectel_networking_subnetpool_v2"
sidebar_current: "docs-selectel-datasource-networking-subnetpool-v2"
description: |-
  Provides information about a subnet pool in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_subnetpool\_v2

Provides information about a subnet pool.

The data source returns exactly one subnet pool. If no pool matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

## Example Usage

```hcl
data "selectel_networking_subnetpool_v2" "subnetpool_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "subnetpool-1"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.

* `region` - (Required) Pool where the subnet pool is located, for example, `ru-3`.

You can combine any of the following selectors. The subnet pool must match all of them at the same time:

* `subnetpool_id` - (Optional) Unique identifier of the subnet pool.

* `name` - (Optional) Name of the subnet pool.

* `description` - (Optional) Description of the subnet pool.

* `default_quota` - (Optional) Maximum number of addresses that can be allocated from the pool.

* `default_prefixlen` - (Optional) Prefix length used for an allocation without an explicit length.

* `min_prefixlen` - (Optional) Smallest prefix length the pool allows.

* `max_prefixlen` - (Optional) Largest prefix length the pool allows.

* `address_scope_id` - (Optional) Unique identifier of the address scope the pool belongs to.

* `ip_version` - (Optional) Internet protocol version of the pool: `4` or `6`.

* `shared` - (Optional) Specifies whether the pool can be used by any project.

* `is_default` - (Optional) Specifies whether the pool is the default one for its address family.

* `tags` - (Optional) Tags that the subnet pool must have.

## Attributes Reference

* `prefixes` - CIDRs the pool allocates from.

* `all_tags` - All tags of the subnet pool.

* `created_at` - Time when the subnet pool was created.

* `updated_at` - Time when the subnet pool was last updated.

* `revision_number` - Revision number of the subnet pool.

~> **Note:** The returned attributes report the actual state of the subnet pool and cannot be used to change it.
