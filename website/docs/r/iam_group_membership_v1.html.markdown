---
layout: "selectel"
page_title: "Selectel: selectel_iam_group_membership_v1"
sidebar_current: "docs-selectel-resource-iam-group_membership-v1"
description: |-
  Creates and manages group membership for Selectel products using public API v1.
---

# selectel\_iam\_group_membership\_v1

Manages group membership for Selectel products using public API v1.
Selectel products support Identity and Access Management (IAM).
For more information about groups, see the [official Selectel documentation](https://docs.selectel.ru/control-panel-actions/users-and-roles/groups/).

## Example Usage

```hcl
resource "selectel_iam_group_membership_v1" "group_membership_1" {
  group_id = selectel_iam_group_v1.group_1.id
  
  user_ids = [
    selectel_iam_user_v1.user_1.keystone_id,
    selectel_iam_serviceuser_v1.serviceuser_1.id
  ]
}
```

## Argument Reference

* `group_id` - (Required) Unique identifier of the group. Retrieved from the [selectel_iam_group_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_group_v1) resource.

* `user_ids` - (Required) List of unique Keystone identifiers of users. Retrieved from the [selectel_iam_serviceuser_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_serviceuser_v1) and [selectel_iam_user_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/iam_user_v1) resources.
