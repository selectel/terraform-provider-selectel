---
layout: "selvpc"
page_title: "SelVPC: selvpc_resell_user_v2"
sidebar_current: "docs-selvpc-resource-resell-user-v2"
description: |-
  Manages a V2 user resource within Resell Selectel VPC.
---

# selvpc\_resell\_user_v2

Manages a V2 user resource within Resell Selectel VPC.

## Example Usage

```hcl
resource "selvpc_resell_user_v2" "user_1" {
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

There is no additional attributes for this resource.

## Import

Users can't be imported at this time.
