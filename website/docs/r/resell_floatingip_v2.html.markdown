---
layout: "selvpc"
page_title: "SelVPC: selvpc_resell_floatingip_v2"
sidebar_current: "docs-selvpc-resource-resell-floatingip-v2"
description: |-
  Manages a V2 floatingip resource within Resell Selectel VPC.
---

# selvpc\_resell\_floatingip_v2

Manages a V2 floatingip resource within Resell Selectel VPC.

## Example Usage

```hcl
resource "selvpc_resell_floatingip_v2" "floatingip_1" {
  project_id = "887e5e35458d4ee38a6ae0543555dec5"
  region     = "ru-1"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) An associated Selectel VPC project. Changing this
  creates a new floating IP.

* `region` - (Required) A region of where the floating IP resides. Changing this
  creates a new floating IP.

## Attributes Reference

The following attributes are exported:

* `port_id` - Contains id of the Networking service port.

* `floating_ip_address` - Contains floating IP address.

* `fixed_ip_address` - Contains internal IP address of the Networking service port.

* `status` - Shows if the license is used or not.

* `servers` - Shows information about servers that use this floating IP. Contains
  `id`, `name` and `status` fields.

## Import

Floating IPs can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selvpc_resell_floatingip_v2.floatingip_1 aa402146-d83e-4c8c-8b74-1f415d4b8253
```
