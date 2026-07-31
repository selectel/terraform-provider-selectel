---
layout: "selectel"
page_title: "Selectel: selectel_networking_router_interface_v2"
sidebar_current: "docs-selectel-resource-networking-router-interface-v2"
description: |-
  Attaches a router to a network in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_router\_interface\_v2

Attaches a router to a network, either through a subnet or through an existing port. The router routes traffic for every subnet attached to it.

## Example Usage

Attaching through a subnet — the API creates the port:

```hcl
resource "selectel_networking_router_interface_v2" "interface_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  router_id  = selectel_networking_router_v2.router_1.id
  subnet_id  = selectel_networking_subnet_v2.subnet_1.id
}
```

Attaching through a port you created yourself:

```hcl
resource "selectel_networking_router_interface_v2" "interface_2" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  router_id  = selectel_networking_router_v2.router_1.id
  port_id    = selectel_networking_port_v2.port_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new interface.

* `region` - (Required) Selectel region where the router is located, for example, `ru-3`. Changing this creates a new interface.

* `router_id` - (Required) Unique identifier of the router the network is attached to. Changing this creates a new interface.

* `subnet_id` - (Optional, Computed) Unique identifier of the subnet to attach. Conflicts with `port_id`. When you attach through `port_id` instead, this is computed from the port. Changing this creates a new interface.

* `port_id` - (Optional, Computed) Unique identifier of the port to attach. Conflicts with `subnet_id`. When you attach through `subnet_id` instead, this is the port the API created for the interface. Changing this creates a new interface.

* `force_destroy` - (Optional) Specifies whether the interface may be detached by first removing the router's static routes that point into the attached subnet. Defaults to `false`. Can be changed without replacement — it only steers this resource's destroy.

~> **Note:** Set exactly one of `subnet_id` and `port_id`. Setting both is rejected at plan; setting neither is rejected before the router is touched.

~> **Note:** A port passed as `port_id` must have exactly one fixed IP. A port with none, or with more than one, is refused before the router is touched, because Neutron cannot derive the interface subnet from it.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is the identifier of the port the router is attached through — whether you supplied that port or the API created it for the subnet.

## Behavior notes

Detaching an interface deletes the port the API created for it. If the interface is detached outside Terraform, or its port is deleted, the resource leaves the state on the next refresh without an error and the next plan re-creates it.

Neutron refuses to detach an interface whose subnet is still in use — for example, when the router has static routes whose `next_hop` is an address inside that subnet. With `force_destroy = false` the detach is retried until the delete timeout and then reported as a diagnostic error, and the resource stays in the state. With `force_destroy = true` those routes — and only those, matched by `next_hop` falling inside the subnet's CIDR — are removed from the router first, and the detach is retried until it succeeds or the timeout expires. Routes pointing anywhere else are never touched.

A delete timeout can be set with a [`timeouts` block](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts); it defaults to 10 minutes.

## Import

You can import an interface by the identifier of its port:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_networking_router_interface_v2.interface_1 <port_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<region>` — Selectel region where the router is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<port_id>` — Unique identifier of the port the router is attached through, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. No composite format and no `router_id` are needed: both `router_id` and `subnet_id` are read from the port.

~> **Note:** If the imported port has more than one fixed IP, `subnet_id` is left unset — there is no single subnet to report. A port that is not attached to a router is rejected as not being a router interface.
