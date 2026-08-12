---
layout: "selectel"
page_title: "Selectel: selectel_vpc_secgroup_v2"
sidebar_current: "docs-selectel-datasource-vpc-secgroup-v2"
description: |-
  Provides information about a security group in the Selectel cloud using VPC API v2.0.
---

# selectel\_vpc\_secgroup\_v2

Provides information about a security group.

The data source returns exactly one security group. If no security group matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html#migrating-data-sources) to replace the corresponding OpenStack data source with the Selectel provider data source.

## Example Usage

```hcl
data "selectel_vpc_secgroup_v2" "secgroup_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "secgroup-1"
}
```

Reading the project's `default` group, to see the rules the platform put in it:

```hcl
data "selectel_vpc_secgroup_v2" "default" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "default"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.

* `region` - (Required) Selectel region where the security group is located, for example, `ru-3`.

You can combine any of the following selectors. The security group must match all of them at the same time:

* `secgroup_id` - (Optional, Computed) Unique identifier of the security group.

* `name` - (Optional, Computed) Name of the security group.

* `description` - (Optional, Computed) Description of the security group.

* `stateful` - (Optional, Computed) Specifies whether the group tracks connection state.

* `tags` - (Optional) Tags that the security group must have.

## Attributes Reference

* `shared` - Specifies whether another project has been granted access to this group.

* `all_tags` - All tags of the security group.

* `security_group_rules` - Identifiers of every rule currently present on the group.
