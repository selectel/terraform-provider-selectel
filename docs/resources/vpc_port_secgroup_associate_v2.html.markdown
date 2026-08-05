---
layout: "selectel"
page_title: "Selectel: selectel_vpc_port_secgroup_associate_v2"
sidebar_current: "docs-selectel-resource-vpc-port-secgroup-associate-v2"
description: |-
  Associates security groups with an existing port in the Selectel cloud using Neutron API v2.0.
---

# selectel\_vpc\_port\_secgroup\_associate\_v2

Associates security groups with an existing port. The resource addresses the port by identifier and never creates or deletes it, nor manages any of its other attributes — use it when the port itself is managed elsewhere, or not by Terraform at all.

## Example Usage

Adding a group to whatever the port already has:

```hcl
resource "selectel_vpc_port_secgroup_associate_v2" "associate_1" {
  project_id         = selectel_vpc_project_v2.project_1.id
  region             = "ru-3"
  port_id            = data.selectel_networking_port_v2.port_1.id
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_1.id]
}
```

Making Terraform the sole authority over the port's groups:

```hcl
resource "selectel_vpc_port_secgroup_associate_v2" "associate_2" {
  project_id         = selectel_vpc_project_v2.project_1.id
  region             = "ru-3"
  port_id            = data.selectel_networking_port_v2.port_1.id
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_1.id]
  enforce            = true
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new resource.

* `region` - (Required) Selectel region where the port is located, for example, `ru-3`. Changing this creates a new resource.

* `port_id` - (Required) Unique identifier of the port. Changing this creates a new resource.

* `security_group_ids` - (Required) Identifiers of the security groups this resource manages on the port. Can be changed without replacement.

* `enforce` - (Optional) Selects the mode, `false` by default. Can be changed without replacement.

## Attributes Reference

* `all_security_group_ids` - The port's complete actual list of security groups, whatever the mode.

## The two modes

**`enforce = false`, the default — additive.** The resource owns only the groups named in `security_group_ids`. Creating it adds them to whatever the port already carries; groups put there by anything else are left alone. An update replaces only the subset the resource previously claimed. Destroying it releases only that subset.

**`enforce = true` — authoritative.** The resource owns the port's entire list. Creating it makes the port's groups exactly `security_group_ids`, removing anything else that was there, however it got there. An update replaces the whole list, and destroying it leaves the port with no security groups at all.

Switching `enforce` from `false` to `true` takes effect on the next apply: the port's groups become exactly the configured set, and anything else is removed even though this resource never added it.

~> **Warning:** Do not manage the same port with both this resource and the `security_group_ids` argument of `selectel_vpc_port_v2`. Both write the port's list, and they will fight over it on every apply. Pick one.

## Behavior notes

`all_security_group_ids` always reports the port's real list, so you can see what else is attached. `security_group_ids` reports what this resource manages: in enforce mode that is the whole list, and in additive mode it is the intersection of what you configured with what the port actually carries — so a group removed externally disappears from it, while a group added externally never appears there.

If the port is deleted outside Terraform, the resource leaves the state on the next refresh without an error, and destroying an association whose port is already gone succeeds without touching anything.

## Import

You can import an association by the identifier of its port:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_vpc_port_secgroup_associate_v2.associate_1 <port_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the port is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<port_id>` — Unique identifier of the port, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`.

~> **Note:** The import fills `all_security_group_ids` with the port's full list, but `security_group_ids` comes out empty and `enforce` takes its default of `false` — there is no configuration yet for the import to reconcile against. Write the configuration you want and check `terraform plan` before applying, especially with `enforce = true`, which would otherwise remove everything you did not list.
