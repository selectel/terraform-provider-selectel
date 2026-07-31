---
layout: "selectel"
page_title: "Selectel: selectel_networking_addressscope_v2"
sidebar_current: "docs-selectel-resource-networking-addressscope-v2"
description: |-
  Creates and manages an address scope in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_addressscope\_v2

Creates and manages an address scope: a namespace subnet pools are attached to, inside which addresses are guaranteed not to overlap.

## Example Usage

```hcl
resource "selectel_networking_addressscope_v2" "addressscope_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "addressscope-1"
  ip_version = 4
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new address scope.

* `region` - (Required) Selectel region where the address scope is created, for example, `ru-3`. Changing this creates a new address scope.

* `name` - (Required) Name of the address scope. Can be changed without replacement.

* `ip_version` - (Optional) Internet protocol version of the address scope, `4` or `6`. Defaults to `4`. Changing this creates a new address scope.

~> **Note:** `ip_version` is not validated locally, so a value other than `4` or `6` is sent to the API and the resulting diagnostic comes from its answer rather than from the plan.

## Attributes Reference

* `shared` - Specifies whether the address scope is shared with other projects.

~> **Note:** `shared` reports the actual state and cannot be used to change it: the cloud's Neutron policy reserves `create_address_scope:shared` and `update_address_scope:shared` for the Keystone admin role, so none of the granular VPC roles can set it. If an administrator shares the scope outside Terraform, the next refresh records that and the plan does not try to undo it. A shared address scope also cannot be made unshared again, whoever asks.

## Behavior notes

An address scope deleted outside Terraform leaves the state on the next refresh without an error, and the next plan re-creates it. A rename made outside Terraform shows up as drift.

An address scope still referenced by a subnet pool through `address_scope_id` is not deleted: the API refuses it with a conflict and the resource stays in the state. Remove or repoint those pools first.

## Import

You can import an address scope:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_networking_addressscope_v2.addressscope_1 <address_scope_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the address scope is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<address_scope_id>` — Unique identifier of the address scope, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`.
