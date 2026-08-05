---
layout: "selectel"
page_title: "Selectel: selectel_vpc_secgroup_rule_v2"
sidebar_current: "docs-selectel-resource-vpc-secgroup-rule-v2"
description: |-
  Creates and manages one rule of a security group in the Selectel cloud using Neutron API v2.0.
---

# selectel\_vpc\_secgroup\_rule\_v2

Creates and manages exactly one rule of an existing security group, leaving every other rule of that group alone.

## Example Usage

Allowing inbound SSH from one subnet:

```hcl
resource "selectel_vpc_secgroup_rule_v2" "rule_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-3"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "192.168.199.0/24"
}
```

Allowing outbound traffic to the members of another group:

```hcl
resource "selectel_vpc_secgroup_rule_v2" "rule_2" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-3"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_1.id
  direction         = "egress"
  ethertype         = "IPv4"
  remote_group_id   = selectel_vpc_secgroup_v2.secgroup_2.id
}
```

## Argument Reference

Every argument is immutable: changing any of them creates a new rule, because a rule has no update path of its own.

* `project_id` - (Required) Unique identifier of the associated project.

* `region` - (Required) Selectel region where the security group is located, for example, `ru-3`.

* `security_group_id` - (Required) Unique identifier of the security group the rule belongs to.

* `direction` - (Required) Direction of the traffic the rule matches, `ingress` or `egress`.

* `ethertype` - (Required) Internet protocol version of the rule, `IPv4` or `IPv6`.

* `description` - (Optional) Description of the rule.

* `protocol` - (Optional, Computed) Protocol the rule matches. Takes either a known name, such as `tcp`, `udp`, `icmp`, `esp` or `ipv6-icmp`, or an IP protocol number as a string. Required when a port range is set.

* `port_range_min` - (Optional, Computed) Lower bound of the port range. Requires `protocol`. With `protocol = "icmp"` this is the ICMP **type**, and it may be set on its own to match one type with any code.

* `port_range_max` - (Optional, Computed) Upper bound of the port range. Requires `protocol` and `port_range_min`. With `protocol = "icmp"` this is the ICMP **code**, which cannot be set without a type.

* `remote_group_id` - (Optional, Computed) Unique identifier of the security group whose members the rule matches. Conflicts with `remote_ip_prefix`.

* `remote_ip_prefix` - (Optional, Computed) CIDR the rule matches. Conflicts with `remote_group_id`.

~> **Note:** `remote_group_id` and `remote_ip_prefix` are mutually exclusive — set at most one. Setting neither means the rule is not restricted by source address.

~> **Note:** `remote_address_group_id` is not supported. Address groups cannot be created by a project member in the Selectel cloud, so there would be nothing for a rule to reference — see the [migration guide](../guides/migrate-from-openstack-provider.html) for the details.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is the identifier the API assigns to the rule, independent of the identifier of its security group and of any other rule on it.

## Behavior notes

The provider checks locally, before calling the API, that the required arguments are present, that the two remote selectors are not both set, that a port range comes with a protocol, and that `port_range_max` is not set without `port_range_min`. That asymmetry mirrors the API: for ICMP the two fields are the type and the code, and a type without a code is valid while the reverse is not. Everything else is validated by the API: `port_range_min` greater than `port_range_max`, a `remote_ip_prefix` whose version disagrees with `ethertype`, an unknown `protocol`, or a `remote_group_id` naming a group that does not exist all fail at apply with the error the API returned, and no rule is created.

Creating a rule whose comparable fields — direction, ethertype, protocol, port range, remote group and remote prefix — exactly match a rule already on the group is refused by the API, whether that existing rule came from Terraform, from the group's own provisioning, or from anywhere else. When comparing prefixes, `0.0.0.0/0`, `::/0` and an absent prefix all count as the same unrestricted value.

`remote_ip_prefix` is stored by Neutron in a canonical, lowercased form. A configuration that writes the same network differently — an IPv6 prefix in upper case, or a host address instead of the network address — does not produce a permanent diff: the provider compares the networks, not the strings.

If the rule is removed outside Terraform, or its whole security group is deleted, the resource leaves the state on the next refresh without an error and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead. Destroying a rule that is already gone succeeds without an error.

## Import

You can import a rule by its own identifier:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_vpc_secgroup_rule_v2.rule_1 <secgroup_rule_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the security group is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<secgroup_rule_id>` — Unique identifier of the rule, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. No composite format is needed: `security_group_id` is read from the rule itself.
