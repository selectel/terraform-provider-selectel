---
layout: "selectel"
page_title: "Selectel: selectel_vpc_port_secgroup_associate_v2"
sidebar_current: "docs-selectel-resource-vpc-port-secgroup-associate-v2"
description: |-
  Creates and manages the association between security groups and an existing port for Selectel products using public API v2.
---

# selectel\_vpc\_port\_secgroup\_associate\_v2

Creates and manages the association between security groups and an existing port using public API v2. For more information about security groups, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/).

The resource addresses the port by identifier, never creates or deletes it, and does not manage any of its other attributes. Use the resource when the port itself is managed elsewhere, or not by Terraform at all.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

### Adding a group to whatever the port already has

```hcl
resource "selectel_vpc_port_secgroup_associate_v2" "associate_1" {
  project_id         = selectel_vpc_project_v2.project_1.id
  region             = "ru-3"
  port_id            = data.selectel_vpc_port_v2.port_1.id
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_1.id]
}
```

### Making Terraform the sole authority over the port's groups

```hcl
resource "selectel_vpc_port_secgroup_associate_v2" "associate_2" {
  project_id         = selectel_vpc_project_v2.project_1.id
  region             = "ru-3"
  port_id            = data.selectel_vpc_port_v2.port_1.id
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_1.id]
  enforce            = true
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new resource. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the port is located, for example, `ru-3`. Changing this creates a new resource. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `port_id` - (Required) Unique identifier of the port. Changing this creates a new resource.

* `security_group_ids` - (Required) Identifiers of the security groups this resource manages on the port. Can be changed without replacement.

* `enforce` - (Optional) Specifies whether `security_group_ids` replaces the port's list of security groups or is appended to it. Boolean flag, the default value is `false`:

  * `false` — the resource owns only the groups listed in `security_group_ids` and leaves the other groups of the port alone.

  * `true` — the resource owns the whole list, so the port keeps only the groups listed in `security_group_ids`.

## Attributes Reference

* `all_security_group_ids` - The port's complete actual list of security groups, whatever the mode.

## Import

You can import an association by the identifier of its port:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_port_secgroup_associate_v2.associate_1 <port_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). The import ID does not contain the project, so `INFRA_PROJECT_ID` must be set.

* `<pool>` — Pool where the port is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/). The import ID does not contain the pool, so `INFRA_REGION` must be set.

* `<port_id>` — Unique identifier of the port, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the port ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ the network page ⟶ the **Ports** tab ⟶ copy the ID of the required port. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack port list` and copy the `ID` field.

~> **Note:** The import fills `all_security_group_ids` with the port's full list, but `security_group_ids` comes out empty and `enforce` takes its default of `false` — there is no configuration yet for the import to reconcile against. Write the configuration you want and check `terraform plan` before applying, especially with `enforce = true`, which would otherwise remove everything you did not list.

## Enforce modes

**`enforce = false`, the default — additive.** The resource owns only the groups named in `security_group_ids`. Creating it adds them to whatever the port already carries; groups put there by anything else are left alone. An update replaces only the subset the resource previously claimed. Destroying it releases only that subset.

**`enforce = true` — authoritative.** The resource owns the port's entire list. Creating it makes the port's groups exactly `security_group_ids`, removing anything else that was there, however it got there. An update replaces the whole list, and destroying it leaves the port with no security groups at all.

Switching `enforce` from `false` to `true` takes effect on the next apply: the port's groups become exactly the configured set, and anything else is removed even though this resource never added it.

~> **Warning:** Do not manage the same port with both this resource and the `security_group_ids` argument of `selectel_vpc_port_v2`. Both write the port's list, and they will fight over it on every apply. Pick one.

## Behavior notes

`all_security_group_ids` always reports the port's real list, so you can see what else is attached. `security_group_ids` reports what this resource manages: in enforce mode that is the whole list, and in additive mode it is the intersection of what you configured with what the port actually carries — so a group removed externally disappears from it, while a group added externally never appears there.

If the port is deleted outside Terraform, the resource leaves the state on the next refresh without an error, and destroying an association whose port is already gone succeeds without touching anything.
