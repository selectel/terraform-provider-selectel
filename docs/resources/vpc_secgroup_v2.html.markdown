---
layout: "selectel"
page_title: "Selectel: selectel_vpc_secgroup_v2"
sidebar_current: "docs-selectel-resource-vpc-secgroup-v2"
description: |-
  Creates and manages a security group for Selectel products using public API v2.
---

# selectel\_vpc\_secgroup\_v2

Creates and manages a security group using public API v2. For more information about security groups, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/).

A security group is a named set of rules that is attached to ports to filter their traffic. Individual rules are managed with [selectel_vpc_secgroup_rule_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_secgroup_rule_v2), and groups are attached to a port with [selectel_vpc_port_secgroup_associate_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_secgroup_associate_v2).

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_secgroup_v2" "secgroup_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-3"
  name        = "secgroup-1"
  description = "Managed by Terraform"
  tags        = ["production", "web"]
}
```

Starting from an empty group, without the egress rules the API adds:

```hcl
resource "selectel_vpc_secgroup_v2" "secgroup_2" {
  project_id           = selectel_vpc_project_v2.project_1.id
  region               = "ru-3"
  name                 = "secgroup-2"
  delete_default_rules = true
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new security group. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the security group is created, for example, `ru-3`. Changing this creates a new security group. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `name` - (Required) Name of the security group. Can be changed without replacement.

* `description` - (Optional) Description of the security group. Removing it from the configuration clears it.

* `stateful` - (Optional) Specifies whether the group tracks connection state. Boolean flag. If you do not specify a value, the API assigns one, and the default value it assigns is `true`. Can be changed without replacement.

* `delete_default_rules` - (Optional) Specifies whether the egress rules that the API adds to a new group are deleted right after the group is created. Boolean flag, the default value is `false`. Changing this creates a new security group, because the flag only takes effect at create time.

* `tags` - (Optional) Tags of the security group that are managed by Terraform.

## Attributes Reference

* `shared` - Specifies whether another project has been granted access to this group.

* `all_tags` - All tags of the security group, including those assigned outside Terraform.

* `security_group_rules` - Identifiers of every rule currently present on the group.

~> **Note:** To share a group, create a [selectel_vpc_rbac_policy_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_rbac_policy_v2) with `action = "access_as_shared"` and `object_type = "security_group"`.

~> **Note:** `security_group_rules` lists every rule actually on the group, whichever resource created it — including rules added by `selectel_vpc_secgroup_rule_v2` and rules the platform adds itself. This resource never reconciles or removes rules it did not create.

## Import

You can import a security group:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_secgroup_v2.secgroup_1 <secgroup_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). The import ID does not contain the project, so `INFRA_PROJECT_ID` must be set.

* `<pool>` — Pool where the security group is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/). The import ID does not contain the pool, so `INFRA_REGION` must be set.

* `<secgroup_id>` — Unique identifier of the security group, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the security group ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Security Groups** tab ⟶ copy the ID of the required security group. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack security group list` and copy the `ID` field.

~> **Note:** `delete_default_rules` describes an action taken at create time and has no counterpart on the API, so it is not filled in by the import. Set it in the configuration to match how you intend the group to be re-created; a different value plans a replacement.

## Behavior notes

If the group is created but a follow-up step fails — removing the default rules, or applying tags — the apply reports the error, and the identifier of the group is already in the state. The group is therefore never left untracked, and the next apply finishes the work instead of creating a second group.

A group deleted outside Terraform leaves the state on the next refresh without an error, and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead.

A group still attached to a port is not deleted. Detach it from every port first.
