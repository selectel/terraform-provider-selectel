---
layout: "selectel"
page_title: "Selectel: selectel_vpc_port_v2"
sidebar_current: "docs-selectel-datasource-vpc-port-v2"
description: |-
  Provides information about a port in Selectel products.
---

# selectel\_vpc\_port\_v2

Provides information about a port. For more information about ports, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/ports/).

The data source returns exactly one port. You can combine any of the optional selectors below, and the port must match all of them at the same time. If no port matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to replace the corresponding OpenStack data source with the Selectel provider data source.

## Example Usage

```hcl
data "selectel_vpc_port_v2" "port_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  fixed_ip   = "192.168.199.50"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the port is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `port_id` - (Optional) Unique identifier of the port. To get the port ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ the network page ⟶ the **Ports** tab ⟶ copy the ID of the required port. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack port list` and copy the `ID` field.

* `name` - (Optional) Name of the port.

* `description` - (Optional) Description of the port.

* `network_id` - (Optional) Unique identifier of the network the port belongs to. To get the network ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ copy the ID of the required network. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack network list` and copy the `ID` field.

* `admin_state_up` - (Optional) Administrative state of the port.

* `device_owner` - (Optional) Owner of the device the port is attached to.

* `device_id` - (Optional) Unique identifier of the device the port is attached to.

* `mac_address` - (Optional) MAC address of the port.

* `status` - (Optional) Status of the port.

* `dns_name` - (Optional) Host name of the port. Private DNS combines it with `dns_domain` of the port's network and creates generated A and AAAA records for the port fixed IPs. Learn more about [Private DNS](https://docs.selectel.ru/en/cloud-servers/private-dns/).

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

* `all_tags` - All tags of the port.

* `port_security_enabled` - Specifies whether port security is enabled.

* `dhcp_blocked` - Specifies whether DHCP is administratively blocked on the port.

* `blocked` - Specifies whether the port is administratively blocked.
