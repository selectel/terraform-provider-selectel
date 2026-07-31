---
layout: "selectel"
page_title: "Selectel: selectel_networking_secgroup_v2"
sidebar_current: "docs-selectel-resource-networking-secgroup-v2"
description: |-
  Creates and manages a security group in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_secgroup\_v2

Creates and manages a security group: a named set of rules that is attached to ports to filter their traffic.

Individual rules are managed with [selectel_networking_secgroup_rule_v2](networking_secgroup_rule_v2.html), and groups are attached to a port with [selectel_networking_port_secgroup_associate_v2](networking_port_secgroup_associate_v2.html).

## Example Usage

```hcl
resource "selectel_networking_secgroup_v2" "secgroup_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-3"
  name        = "secgroup-1"
  description = "Managed by Terraform"
  tags        = ["production", "web"]
}
```

Starting from an empty group, without the egress rules the API adds:

```hcl
resource "selectel_networking_secgroup_v2" "secgroup_2" {
  project_id           = selectel_vpc_project_v2.project_1.id
  region               = "ru-3"
  name                 = "secgroup-2"
  delete_default_rules = true
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new security group.

* `region` - (Required) Selectel region where the security group is created, for example, `ru-3`. Changing this creates a new security group.

* `name` - (Required) Name of the security group. Can be changed without replacement.

* `description` - (Optional) Description of the security group. Removing it from the configuration clears it.

* `stateful` - (Optional, Computed) Specifies whether the group tracks connection state. Defaults to the value the API assigns, which is `true`. Can be changed without replacement.

* `delete_default_rules` - (Optional) Specifies whether the rules the API creates together with the group — one egress rule per supported ethertype — are removed right after it is created. Defaults to `false`. Changing this creates a new security group, because it only takes effect at create time.

* `tags` - (Optional) Tags of the security group that are managed by Terraform.

## Attributes Reference

* `shared` - Specifies whether another project has been granted access to this group.

* `all_tags` - All tags of the security group, including those assigned outside Terraform.

* `security_group_rules` - Identifiers of every rule currently present on the group.

~> **Note:** `shared` reports the actual state and cannot be used to change it — the attribute is read-only on the API for every caller, administrators included. To share a group, create a [selectel_networking_rbac_policy_v2](networking_rbac_policy_v2.html) with `action = "access_as_shared"` and `object_type = "security_group"`.

~> **Note:** `security_group_rules` lists every rule actually on the group, whichever resource created it — including rules added by `selectel_networking_secgroup_rule_v2` and rules the platform adds itself. This resource never reconciles or removes rules it did not create.

## Behavior notes

In the `default` security group of a project, in a region where the platform enables its all-pass flag, the rules created when the group was first provisioned include an extra allow-all ingress rule per ethertype, added by the platform rather than by Terraform. This is a regional setting that the API does not expose, so the provider cannot predict it — it only reports the rules through `security_group_rules`. Their presence is not an error and produces no permanent diff.

If the group is created but a follow-up step fails — removing the default rules, or applying tags — the apply reports the error, and the identifier of the group is already in the state. The group is therefore never left untracked, and the next apply finishes the work instead of creating a second group.

A group deleted outside Terraform leaves the state on the next refresh without an error, and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead.

A group still attached to a port is not deleted: the API refuses it with a conflict and the resource stays in the state. Detach it from every port first. Separately, deleting the `default` group of a project is refused without cloud-administrator privileges — none of the granular VPC roles lifts that protection.

## Import

You can import a security group:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_networking_secgroup_v2.secgroup_1 <secgroup_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the security group is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<secgroup_id>` — Unique identifier of the security group, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`.

~> **Note:** `delete_default_rules` describes an action taken at create time and has no counterpart on the API, so it is not filled in by the import. Set it in the configuration to match how you intend the group to be re-created; a different value plans a replacement.
