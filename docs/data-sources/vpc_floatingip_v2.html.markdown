---
layout: "selectel"
page_title: "Selectel: selectel_vpc_floatingip_v2"
sidebar_current: "docs-selectel-datasource-vpc-floatingip-v2"
description: |-
  Provides information about a floating IP in the Selectel cloud using VPC API v2.0.
---

# selectel\_vpc\_floatingip\_v2

Provides information about a floating IP.

The data source returns exactly one floating IP. If no floating IP matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html#migrating-data-sources) to replace the corresponding OpenStack data source with the Selectel provider data source.

## Example Usage

```hcl
data "selectel_vpc_floatingip_v2" "floatingip_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  address    = "203.0.113.10"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.

* `region` - (Required) Selectel region where the floating IP is located, for example, `ru-3`.

You can combine any of the following selectors. The floating IP must match all of them at the same time:

* `floatingip_id` - (Optional, Computed) Unique identifier of the floating IP.

* `address` - (Optional, Computed) Public address of the floating IP.

* `description` - (Optional, Computed) Description of the floating IP.

* `pool` - (Optional) Name of the external network the address is allocated from. The provider resolves the name to the identifier of the network before the search, so an unknown or ambiguous name fails with a diagnostic error.

* `port_id` - (Optional, Computed) Unique identifier of the port the floating IP is associated with.

* `fixed_ip` - (Optional, Computed) Address of the associated port traffic is forwarded to.

* `status` - (Optional, Computed) Status of the floating IP, for example, `ACTIVE` or `DOWN`.

* `tags` - (Optional) Tags that the floating IP must have.

## Attributes Reference

* `floating_network_id` - Unique identifier of the external network the address is allocated from.

* `router_id` - Unique identifier of the router that routes the address.

* `all_tags` - All tags of the floating IP.

* `qos_policy_id` - Unique identifier of the QoS policy associated with the floating IP.

* `blocked` - Specifies whether the floating IP is administratively blocked.
