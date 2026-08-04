---
layout: "selectel"
page_title: "Selectel: selectel_networking_floatingip_associate_v2"
sidebar_current: "docs-selectel-resource-networking-floatingip-associate-v2"
description: |-
  Associates an existing floating IP with an existing port in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_floatingip\_associate\_v2

Associates an existing floating IP with an existing port. This resource manages only the association: it does not create or delete the floating IP or the port.

## Example Usage

```hcl
resource "selectel_networking_floatingip_associate_v2" "association_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-3"
  floating_ip = selectel_networking_floatingip_v2.floatingip_1.address
  port_id     = selectel_networking_port_v2.port_1.id
  fixed_ip    = "192.168.199.50"

  # The subnet of the port must already have a route to the external network the
  # floating IP comes from, and no argument of this resource refers to the router
  # interface that provides it.
  depends_on = [selectel_networking_router_interface_v2.interface_1]
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new association.

* `region` - (Required) Selectel region where the floating IP and port are located, for example, `ru-3`. Changing this creates a new association.

* `floating_ip` - (Required) Address of the existing floating IP to associate. Changing this creates a new association.

* `port_id` - (Required) Unique identifier of the existing port. Changing this updates the association in place.

* `fixed_ip` - (Optional, Computed) Fixed IP address of `port_id` to associate. If omitted, the remote API selects an address and the provider stores it in state. Changing it updates the association in place.

## Attributes Reference

No attributes are exported beyond the arguments. The resource identifier is the remote identifier of the floating IP, not its address.

## Behavior notes

Changes made outside Terraform are reflected during refresh. If the association is detached or moved, the next plan restores the configured port and fixed IP. If the floating IP itself is deleted, the association leaves state.

Deleting this resource detaches the floating IP without deleting either the floating IP or the port. Detaching an already detached floating IP is idempotent. If the floating IP itself no longer exists when the resource is deleted, the provider reports an error instead of claiming that it detached the association.

## Import

You can import an association by the remote identifier of its floating IP:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_networking_floatingip_associate_v2.association_1 <floatingip_id>
```

The import reads `project_id` and `region` from the provider environment and fills `floating_ip`, `port_id`, and `fixed_ip` from the remote floating IP.
