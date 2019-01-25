---
layout: "selectel"
page_title: "Selectel: selectel_vpc_role_v2"
sidebar_current: "docs-selectel-resource-vpc-role-v2"
description: |-
  Manages a V2 role resource within Selectel VPC.
---

# selectel\_vpc\_role_v2

Manages a V2 role resource within Selectel VPC.

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
}

resource "selectel_vpc_user_v2" "user_1" {
  password    = "secret"
}

resource "selectel_vpc_role_v2" "role_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_1.id}"
  user_id    = "${selectel_vpc_user_v2.user_1.id}"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) An associated Selectel VPC project. Changing this
  creates a new role.

* `user_id` - (Required) An associated Selectel VPC user. Changing this
  creates a new role.

## Attributes Reference

There are no additional attributes for this resource.

## Import

Roles can be imported by specifying `project_id` and `user_id` arguments,
separated by a forward slash:

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selectel_vpc_role_v2.role_1 <project_id>/<user_id>
```
