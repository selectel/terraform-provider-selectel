---
layout: "selectel"
page_title: "Selectel: selectel_networking_port_forwarding_v2"
sidebar_current: "docs-selectel-resource-networking-port-forwarding-v2"
description: |-
  Creates and manages a port-forwarding rule of a floating IP in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_port\_forwarding\_v2

Creates and manages one port-forwarding rule of a floating IP: traffic arriving at an external port of the floating IP is forwarded to an internal address and port.

## Example Usage

```hcl
resource "selectel_networking_port_forwarding_v2" "pf_1" {
  project_id          = selectel_vpc_project_v2.project_1.id
  region              = "ru-3"
  floatingip_id       = selectel_networking_floatingip_v2.floatingip_1.id
  internal_port_id    = selectel_networking_port_v2.port_1.id
  internal_ip_address = "192.168.199.50"
  internal_port       = 80
  external_port       = 8080
  protocol            = "tcp"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new rule.

* `region` - (Required) Selectel region where the floating IP is located, for example, `ru-3`. Changing this creates a new rule.

* `floatingip_id` - (Required) Unique identifier of the floating IP the rule belongs to. Changing this creates a new rule, because the rule exists under one floating IP.

* `internal_port_id` - (Required) Unique identifier of the port traffic is forwarded to. Can be changed without replacement.

* `internal_ip_address` - (Required) Address traffic is forwarded to, which must be one of the fixed IPs of `internal_port_id`. Can be changed without replacement.

* `internal_port` - (Required) Port traffic is forwarded to. Can be changed without replacement.

* `external_port` - (Required) Port of the floating IP traffic arrives at. Can be changed without replacement.

* `protocol` - (Required) Protocol of the forwarded traffic, for example, `tcp` or `udp`. Can be changed without replacement.

* `description` - (Optional) Description of the rule.

~> **Note:** The provider does no local validation of `protocol`, `internal_port` and `external_port`: the values are validated by the API, so an invalid value fails at apply rather than at plan. A rule whose `external_port` and `protocol` pair is already used by another rule of the same floating IP is refused with a diagnostic error.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is the identifier the API assigns to the rule.

## Behavior notes

A rule removed outside Terraform, or one whose floating IP was deleted, leaves the state on the next refresh without an error, and the next plan re-creates it — the rule cannot exist without its floating IP. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead.

## Import

You can import a rule by the identifier of its floating IP and the identifier of the rule:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_networking_port_forwarding_v2.pf_1 <floatingip_id>/<id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the floating IP is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<floatingip_id>` — Unique identifier of the floating IP, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`.

* `<id>` — Unique identifier of the rule, for example, `9f8e7d6c-5b4a-3928-1706-f5e4d3c2b1a0`.

An identifier without exactly one `/` separator and two non-empty parts is rejected locally without calling the API.
