---
layout: "selvpc"
page_title: "SelVPC: selvpc_resell_token_v2"
sidebar_current: "docs-selvpc-resource-resell-token-v2"
description: |-
  Manages a V2 token resource within Resell Selectel VPC.
---

# selvpc\_resell\_token_v2

Manages a V2 token resource within Resell Selectel VPC.

ID of this resource can be used within the OpenStack API Identity service as
the `X-Auth-Token` value.

## Example Usage

```hcl
resource "selvpc_resell_project_v2" "project_1" {
  auto_quotas = true
}

resource "selvpc_resell_token_v2" "token_tf_acc_test_1" {
  project_id = "${selvpc_resell_project_v2.project_1.id}"
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
