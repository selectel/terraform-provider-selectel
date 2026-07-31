---
layout: "selectel"
page_title: "Selectel: selectel_networking_rbac_policy_v2"
sidebar_current: "docs-selectel-resource-networking-rbac-policy-v2"
description: |-
  Creates and manages an RBAC policy in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_rbac\_policy\_v2

Creates and manages an RBAC policy: it grants another project access to one of your networking objects — for example, sharing a network so that project can create ports on it.

## Example Usage

```hcl
resource "selectel_networking_rbac_policy_v2" "rbac_1" {
  project_id    = selectel_vpc_project_v2.project_1.id
  region        = "ru-3"
  action        = "access_as_shared"
  object_type   = "network"
  object_id     = selectel_networking_network_v2.network_1.id
  target_tenant = selectel_vpc_project_v2.project_2.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the project that owns the policy. Changing this creates a new policy.

* `region` - (Required) Selectel region where the object is located, for example, `ru-3`. Changing this creates a new policy.

* `action` - (Required) Access granted by the policy. One of `access_as_shared` or `access_as_external`; any other value is rejected at plan. Changing this creates a new policy.

* `object_type` - (Required) Type of the object being shared. One of `network`, `qos_policy`, `security_group`, `subnetpool`, `address_scope`, `address_group` or `bgpvpn`; any other value is rejected at plan. Changing this creates a new policy.

* `object_id` - (Required) Unique identifier of the object being shared. Changing this creates a new policy.

* `target_tenant` - (Required) Unique identifier of the project the access is granted to. Can be changed without replacement, which moves the grant from one project to another.

~> **Note:** `qos_policy` and `bgpvpn` are accepted even though this provider has no resource for either: a policy can reference an existing object of that type by its identifier.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is the identifier the API assigns to the policy.

## Behavior notes

`action` and `object_type` are each validated locally against a closed set, so a typo fails at plan. Their *combination* is validated by the API — `access_as_external` does not apply to every object type — so an unsupported pair fails at apply, and nothing is created.

Creating, changing or deleting a policy requires you to be a member of the owning project, or a cloud administrator. The narrower VPC roles — `vpc.private_network.admin`, `vpc.external_access.admin`, `vpc.network_security.admin` and their viewer and user variants — are not sufficient on their own, and a caller without a sufficient role gets an authorization error from the API without anything changing. A wildcard `target_tenant`, which shares the object with every project, is accepted only for a cloud administrator.

A policy deleted outside Terraform leaves the state on the next refresh without an error, and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead — losing the role needed to read the policy is not the same as the policy being gone.

Deleting a policy the target project still depends on — it has ports or subnets relying on the shared object — is refused with a conflict, and the resource stays in the state. Remove those dependent resources first.

## Import

You can import a policy:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_networking_rbac_policy_v2.rbac_1 <rbac_policy_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the project that owns the policy. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the object is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<rbac_policy_id>` — Unique identifier of the policy, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`.
