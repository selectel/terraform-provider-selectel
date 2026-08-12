---
layout: "selectel"
page_title: "Selectel: selectel_vpc_port_v2"
sidebar_current: "docs-selectel-resource-vpc-port-v2"
description: |-
  Creates and manages a port in the Selectel cloud using VPC API v2.0.
---

# selectel\_vpc\_port\_v2

Creates and manages a port using VPC API v2.0.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_port_v2" "port_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  network_id = selectel_vpc_network_v2.network_1.id
  name       = "port-1"

  fixed_ip {
    subnet_id  = selectel_vpc_network_subnet_v2.subnet_1.id
    ip_address = "192.168.199.50"
  }

  allowed_address_pairs {
    ip_address = "192.168.199.100"
  }

  binding {
    vnic_type = "normal"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new port.

* `region` - (Required) Pool where the port is located, for example, `ru-3`. Changing this creates a new port.

* `network_id` - (Required) Unique identifier of the network the port belongs to. Changing this creates a new port.

* `name` - (Optional) Name of the port. If you do not specify a name, the API assigns one.

* `description` - (Optional) Description of the port.

* `admin_state_up` - (Optional) Administrative state of the port. If you do not specify the value, the API assigns one.

* `mac_address` - (Optional) MAC address of the port.

* `device_id` - (Optional) Unique identifier of the device the port is attached to. Changing this creates a new port.

* `device_owner` - (Optional) Owner of the device the port is attached to. Changing this creates a new port.

* `fixed_ip` - (Optional) Fixed IPs of the port. Conflicts with `no_fixed_ip`. If you specify neither, the API allocates an address from a DHCP subnet of the network. You can specify several blocks:

  * `subnet_id` - (Required) Unique identifier of the subnet the address is allocated from.

  * `ip_address` - (Optional) Specific address to assign. If you do not specify it, the API allocates one from the subnet.

* `no_fixed_ip` - (Optional) Creates the port without any fixed IP, even when the network has a DHCP subnet. The default value is `false`. Conflicts with `fixed_ip`.

* `security_group_ids` - (Optional) Security groups assigned to the port. Conflicts with `no_security_groups`. If you specify neither, the port gets the project's default security group.

* `no_security_groups` - (Optional) Creates the port without any security group. The default value is `false`. Conflicts with `security_group_ids`.

* `allowed_address_pairs` - (Optional) Additional addresses the port may send traffic from. You can specify several blocks:

  * `ip_address` - (Required) Address or CIDR of the pair.

  * `mac_address` - (Optional) MAC address of the pair. If you do not specify it, the port's own MAC address is used.

* `extra_dhcp_option` - (Optional) Extra DHCP options served to the port. You can specify several blocks:

  * `name` - (Required) Name of the option.

  * `value` - (Required) Value of the option.

  * `ip_version` - (Optional) Internet protocol version of the option: `4` (default) or `6`.

* `binding` - (Optional) Binding parameters of the port. At most one block:

  * `vnic_type` - (Optional) Virtual NIC type: `normal`, `direct`, `direct-physical`, `macvtap`, `baremetal`, `virtio-forwarder`, or `remote-managed`. Any other value is rejected before the API is called.

* `dns_name` - (Optional) DNS name of the port.

* `tags` - (Optional) Tags of the port that Terraform manages. Tags added outside Terraform are reported in `all_tags` and are not removed by this resource.

## Attributes Reference

* `all_fixed_ips` - Fixed IPs the port actually has, which may differ from `fixed_ip` when the API allocates addresses:

  * `subnet_id` - Unique identifier of the subnet.

  * `ip_address` - Assigned address.

* `all_security_group_ids` - Security groups the port actually has, which may differ from `security_group_ids`.

* `all_tags` - All tags of the port, including the ones added outside Terraform.

* `status` - Status of the port.

* `dns_assignment` - DNS records assigned to the port:

  * `ip_address` - Address the record points to.

  * `hostname` - Host name of the record.

  * `fqdn` - Fully-qualified domain name of the record.

* `port_security_enabled` - Specifies whether port security is enabled.

* `qos_policy_id` - Unique identifier of the QoS policy associated with the port.

* `dhcp_blocked` - Specifies whether DHCP is administratively blocked on the port.

* `blocked` - Specifies whether the port is administratively blocked.

## Import

You can import a port:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_port_v2.port_1 <port_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<pool>` — Pool where the port is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<port_id>` — Unique identifier of the port, for example, `9d1e2f3a-4b5c-6d70-8e9f-0a1b2c3d4e5f`.
