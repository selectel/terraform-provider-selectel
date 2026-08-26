---
layout: "selectel"
page_title: "Selectel: selectel_vpc_secgroup_rule_v2"
sidebar_current: "docs-selectel-resource-vpc-secgroup-rule-v2"
description: |-
  Creates and manages a rule of a security group for Selectel products using public API v2.
---

# selectel\_vpc\_secgroup\_rule\_v2

Creates and manages a rule of a security group using public API v2. For more information about security groups, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/).

The resource manages exactly one rule of an existing security group and leaves every other rule of that group alone. The group itself is managed with the [selectel_vpc_secgroup_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_secgroup_v2) resource.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

### Allowing inbound SSH from one subnet

```hcl
resource "selectel_vpc_secgroup_rule_v2" "rule_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-6"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "192.168.199.0/24"
}
```

### Allowing outbound traffic to the members of another group

```hcl
resource "selectel_vpc_secgroup_rule_v2" "rule_2" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-6"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_1.id
  direction         = "egress"
  ethertype         = "IPv4"
  remote_group_id   = selectel_vpc_secgroup_v2.secgroup_2.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new rule. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the security group is located, for example, `ru-6`. Changing this creates a new rule. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `security_group_id` - (Required) Unique identifier of the security group the rule belongs to. Changing this creates a new rule. Retrieved from the [selectel_vpc_secgroup_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_secgroup_v2) resource.

* `direction` - (Required) Direction of the traffic the rule matches, `ingress` or `egress`. Changing this creates a new rule.

* `ethertype` - (Required) Internet protocol version of the rule, `IPv4` or `IPv6`. Changing this creates a new rule.

* `description` - (Optional) Description of the rule. Changing this creates a new rule.

* `protocol` - (Optional) Protocol the rule matches. Changing this creates a new rule. Takes either a known name, such as `tcp`, `udp`, `icmp`, `esp` or `ipv6-icmp`, or an IP protocol number as a string. Required when a port range is set.

* `port_range_min` - (Optional) Lower bound of the port range. Changing this creates a new rule. Requires `protocol`. With `protocol = "icmp"` this is the ICMP **type**, and it may be set on its own to match one type with any code.

* `port_range_max` - (Optional) Upper bound of the port range. Changing this creates a new rule. Requires `protocol` and `port_range_min`. With `protocol = "icmp"` this is the ICMP **code**, which cannot be set without a type.

* `remote_group_id` - (Optional) Unique identifier of the security group whose members the rule matches. Changing this creates a new rule. Retrieved from the [selectel_vpc_secgroup_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/vpc_secgroup_v2) data source.

* `remote_ip_prefix` - (Optional) CIDR the rule matches. Changing this creates a new rule. Conflicts with `remote_group_id`. If you specify an address with host bits set, for example 192.168.1.42/24, the rule is stored with the network address of the prefix — 192.168.1.0/24. The provider does not report this normalization as a configuration drift.

~> **Note:** `remote_group_id` and `remote_ip_prefix` are mutually exclusive — set at most one. Setting neither means the rule is not restricted by remote address — the source for ingress rules, the destination for egress rules.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is the identifier the API assigns to the rule, independent of the identifier of its security group and of any other rule on it.

## Import

You can import a rule by its own identifier:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_secgroup_rule_v2.rule_1 <secgroup_rule_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/user-types).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `<pool>` — Pool where the security group is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `<secgroup_rule_id>` — Unique identifier of the rule, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the rule ID, use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack security group rule list` and copy the `ID` field. The Control panel does not show the rule ID.

## Behavior notes

If the rule is removed outside Terraform, or its whole security group is deleted, the resource leaves the state on the next refresh without an error and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead. Destroying a rule that is already gone succeeds without an error.
