---
layout: "selectel"
page_title: "Selectel: selectel_vpc_rbac_policy_v2"
sidebar_current: "docs-selectel-resource-vpc-rbac-policy-v2"
description: |-
  Creates and manages an RBAC policy for Selectel products using public API v2.
---

# selectel\_vpc\_rbac\_policy\_v2

Creates and manages an RBAC policy that grants another project access to a networking object using public API v2. For more information about managing network access, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/manage-access/).

For example, share a network with another project so that project can create ports on it.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_rbac_policy_v2" "rbac_1" {
  project_id    = selectel_vpc_project_v2.project_1.id
  region        = "ru-3"
  action        = "access_as_shared"
  object_type   = "network"
  object_id     = selectel_vpc_network_v2.network_1.id
  target_tenant = selectel_vpc_project_v2.project_2.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the project that owns the policy. Changing this creates a new policy. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the object is located, for example, `ru-3`. Changing this creates a new policy. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `action` - (Required) Access granted by the policy. Available values are `access_as_shared` and `access_as_external`. `access_as_external` is applicable only when `object_type` is `network`. Changing this creates a new policy.

* `object_type` - (Required) Type of the object to share. Available values are `network` and `security_group`. Changing this creates a new policy.

* `object_id` - (Required) Unique identifier of the object being shared. Changing this creates a new policy.

* `target_tenant` - (Required) Unique identifier of the project the access is granted to. Can be changed without replacement, which moves the grant from one project to another.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is the identifier the API assigns to the policy.

## Import

You can import a policy:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_rbac_policy_v2.rbac_1 <rbac_policy_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). The import ID does not contain the project, so `INFRA_PROJECT_ID` must be set.

* `<pool>` — Pool where the object is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/). The import ID does not contain the pool, so `INFRA_REGION` must be set.

* `<rbac_policy_id>` — Unique identifier of the policy, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the policy ID, use [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack network rbac list` and copy the `ID` field. The Control panel does not show the policy ID.

## Behavior notes

A policy deleted outside Terraform leaves the state on the next refresh without an error, and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead — losing the role needed to read the policy is not the same as the policy being gone.

Deleting a policy the target project still depends on — it has ports or subnets relying on the shared object — is refused with a conflict, and the resource stays in the state. Remove those dependent resources first.
