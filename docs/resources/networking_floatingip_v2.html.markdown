---
layout: "selectel"
page_title: "Selectel: selectel_networking_floatingip_v2"
sidebar_current: "docs-selectel-resource-networking-floatingip-v2"
description: |-
  Creates and manages a floating IP in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_floatingip\_v2

Creates and manages a floating IP: a public address allocated from an external network that can be associated with a port and moved between ports without being re-allocated.

## Example Usage

```hcl
resource "selectel_networking_floatingip_v2" "floatingip_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-3"
  pool        = "external-network"
  description = "Address of the web server"
  tags        = ["production", "web"]

  port_id  = selectel_networking_port_v2.port_1.id
  fixed_ip = "192.168.199.50"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new floating IP.

* `region` - (Required) Selectel region where the floating IP is created, for example, `ru-3`. Changing this creates a new floating IP.

* `pool` - (Required) Name of the external network the address is allocated from. The provider resolves the name to the identifier of the network before the address is allocated. Changing this creates a new floating IP.

* `subnet_id` - (Optional) Unique identifier of a subnet of the external network the address must be allocated from. Use it when the external network has more than one subnet and the address must come from a specific one. Changing this creates a new floating IP.

* `description` - (Optional) Description of the floating IP.

* `port_id` - (Optional, Computed) Unique identifier of the port the floating IP is associated with. Can be changed without replacement, which moves the address between ports. Removing it detaches the address and keeps it allocated to the project.

* `fixed_ip` - (Optional, Computed) Address of `port_id` traffic is forwarded to, which must be one of the fixed IPs of that port. Use it when the port has more than one address. Can be changed without replacement.

* `dns_name` - (Optional, Computed) DNS name of the floating IP. Changing this creates a new floating IP.

* `dns_domain` - (Optional, Computed) DNS domain of the floating IP. A fully-qualified (unambiguous) name must have a dot at the end, for example, `example.com.` — a value without it is rejected at plan. Changing this creates a new floating IP.

* `tags` - (Optional) Tags of the floating IP that are managed by Terraform.

~> **Note:** A specific address cannot be requested: the API always assigns one from `pool`. To keep an address, keep the resource — the address is released only when the floating IP is deleted.

## Attributes Reference

* `address` - Public address assigned by the API.

* `floating_network_id` - Unique identifier of the external network `pool` resolved to.

* `router_id` - Unique identifier of the router that routes the address, assigned by the API when the floating IP is associated with a port.

* `status` - Status of the floating IP, for example, `ACTIVE` or `DOWN`.

* `all_tags` - All tags of the floating IP, including those assigned outside Terraform.

* `qos_policy_id` - Unique identifier of the QoS policy associated with the floating IP.

* `blocked` - Specifies whether the floating IP is administratively blocked.

~> **Note:** `qos_policy_id` and `blocked` report the actual state and cannot be used to change it: the cloud's Neutron policy reserves both for the Keystone admin role.

## Behavior notes

An unknown `pool`, or one that matches more than one external network, fails before any address is allocated, and the two cases are reported as distinct errors.

A floating IP deleted outside Terraform leaves the state on the next refresh without an error, and the next plan re-creates it — with a new address, because the address cannot be requested. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead.

## Import

You can import a floating IP:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_networking_floatingip_v2.floatingip_1 <floatingip_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the floating IP is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<floatingip_id>` — Unique identifier of the floating IP, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`.

~> **Note:** `subnet_id` is not filled in by the import either: the API accepts it when the address is allocated but never returns it, so the provider cannot read it back. A configured `subnet_id` against an imported floating IP does **not** plan a replacement — the difference is treated as an unknown value rather than a change, so the allocated address is never released over it.

~> **Note:** The API returns the identifier of the external network, not its name, so `pool` is not filled in by the import. Set it in the configuration to the name of the network `floating_network_id` refers to; a different name plans a replacement.
