---
layout: "selectel"
page_title: "Selectel: selectel_vpc_token_v2"
sidebar_current: "docs-selectel-resource-vpc-token-v2"
description: |-
  Manages a V2 token resource within Selectel VPC.
---

# selectel\_vpc\_token_v2

Manages a V2 token resource within Selectel VPC.

ID of this resource can be used within the OpenStack API Identity service as
the `X-Auth-Token` value.

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
}

resource "selectel_vpc_token_v2" "token_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_1.id}"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Optional) An associated Selectel VPC project. Changing this
  creates a new token.

* `account_name` - (Optional) An associated Selectel VPC account. Changing this
  creates a new token.

## Attributes Reference

There are no additional attributes for this resource.

## Import

Tokens can't be imported at this time.
