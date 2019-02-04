---
layout: "selectel"
page_title: "Selectel: selectel_vpc_user_v2"
sidebar_current: "docs-selectel-resource-vpc-user-v2"
description: |-
  Manages a V2 user resource within Selectel VPC.
---

# selectel\_vpc\_user_v2

Manages a V2 user resource within Selectel VPC.

## Example Usage

```hcl
resource "selectel_vpc_user_v2" "user_1" {
  password = "verysecret"
  enabled  = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the user. Changing this updates the name of the
  existing user.

* `password` - (Required) Password of the user. Changing this updates the
  password of the existing user.

* `enabled` - (Optional) Enabled state of the user. Changing this updates the
  enabled state of the existing user.

## Attributes Reference

There are no additional attributes for this resource.

## Import

Users can't be imported at this time.
