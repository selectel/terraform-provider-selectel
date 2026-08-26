---
layout: "selectel"
page_title: "Selectel: selectel_vpc_secgroup_v2"
sidebar_current: "docs-selectel-datasource-vpc-secgroup-v2"
description: |-
  Provides information about a security group in Selectel products.
---

# selectel\_vpc\_secgroup\_v2

Provides information about a security group. For more information about security groups, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/).

The data source returns exactly one security group. You can combine any of the optional selectors below, and the security group must match all of them at the same time. If no security group matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to replace the corresponding OpenStack data source with the Selectel provider data source.

## Example Usage

### Find the group by name

```hcl
data "selectel_vpc_secgroup_v2" "secgroup_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name       = "secgroup-1"
}
```

### Find the default group of the project

```hcl
data "selectel_vpc_secgroup_v2" "default" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name       = "default"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the security group is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `secgroup_id` - (Optional) Unique identifier of the security group.

* `name` - (Optional) Name of the security group.

* `description` - (Optional) Description of the security group.

* `stateful` - (Optional) Specifies whether the group tracks connection state.

* `tags` - (Optional) Tags that the security group must have.

## Attributes Reference

* `shared` - Specifies whether another project has been granted access to this group.

* `all_tags` - All tags of the security group.

* `security_group_rules` - Identifiers of every rule currently present on the group.
