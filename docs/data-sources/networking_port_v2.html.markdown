---
layout: "selectel"
page_title: "Selectel: selectel_networking_port_v2"
sidebar_current: "docs-selectel-datasource-networking-port-v2"
description: |-
  Provides information about a port in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_port\_v2

Provides information about a port.

The data source returns exactly one port. If no port matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

## Example Usage

```hcl
data "selectel_networking_port_v2" "port_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  fixed_ip   = "192.168.199.50"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.

* `region` - (Required) Pool where the port is located, for example, `ru-3`.

You can combine any of the following selectors. The port must match all of them at the same time:

* `port_id` - (Optional) Unique identifier of the port.

* `name` - (Optional) Name of the port.

* `description` - (Optional) Description of the port.

* `network_id` - (Optional) Unique identifier of the network the port belongs to.

* `admin_state_up` - (Optional) Administrative state of the port.

* `device_owner` - (Optional) Owner of the device the port is attached to.

* `device_id` - (Optional) Unique identifier of the device the port is attached to.

* `mac_address` - (Optional) MAC address of the port.

* `status` - (Optional) Status of the port.

* `dns_name` - (Optional) DNS name of the port.

* `fixed_ip` - (Optional) Single address that must be among the port's fixed IPs. The port matches when any of its addresses equals this value, not only when it is the port's sole address.

* `security_group_ids` - (Optional) Security groups of which the port must have at least one assigned. The port does not need the exact same set.

* `tags` - (Optional) Tags that the port must have.

## Attributes Reference

* `all_fixed_ips` - Fixed IPs of the port:

  * `subnet_id` - Unique identifier of the subnet.

  * `ip_address` - Assigned address.

* `all_security_group_ids` - Security groups assigned to the port.

* `allowed_address_pairs` - Additional addresses the port may send traffic from:

  * `ip_address` - Address or CIDR of the pair.

  * `mac_address` - MAC address of the pair.

* `extra_dhcp_option` - Extra DHCP options served to the port:

  * `name` - Name of the option.

  * `value` - Value of the option.

  * `ip_version` - Internet protocol version of the option.

* `binding` - Binding parameters of the port:

  * `vnic_type` - Virtual NIC type of the port.

* `dns_assignment` - DNS records assigned to the port:

  * `ip_address` - Address the record points to.

  * `hostname` - Host name of the record.

  * `fqdn` - Fully-qualified domain name of the record.

* `all_tags` - All tags of the port.

* `port_security_enabled` - Specifies whether port security is enabled.

* `qos_policy_id` - Unique identifier of the QoS policy associated with the port.

* `dhcp_blocked` - Specifies whether DHCP is administratively blocked on the port.

* `blocked` - Specifies whether the port is administratively blocked.

~> **Note:** The returned attributes report the actual state of the port and cannot be used to change it. `binding.host_id` and `binding.profile` are not exposed: the cloud's Neutron policy reserves them for the Keystone admin role and the networking SDK does not return them.
