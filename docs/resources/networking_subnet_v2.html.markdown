---
layout: "selectel"
page_title: "Selectel: selectel_networking_subnet_v2"
sidebar_current: "docs-selectel-resource-networking-subnet-v2"
description: |-
  Creates and manages a subnet in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_subnet\_v2

Creates and manages a subnet using Neutron API v2.0. For more information about networks and subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/servers/networks/about-networks/).

~> **Note:** `service_types` and `blocked` are writable only by the Keystone admin role, which no granular Selectel VPC role grants, so this resource reports them as read-only attributes. `segment_id` is not available at all. If you are migrating a configuration from `openstack_networking_subnet_v2`, remove those arguments.

## Example Usage

```hcl
resource "selectel_networking_subnet_v2" "subnet_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  network_id = selectel_networking_network_v2.network_1.id
  cidr       = "192.168.199.0/24"
  name       = "subnet-1"

  dns_nameservers = ["8.8.8.8", "1.1.1.1"]

  allocation_pool {
    start = "192.168.199.10"
    end   = "192.168.199.100"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new subnet. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource.

* `region` - (Required) Pool where the subnet is located, for example, `ru-3`. Changing this creates a new subnet.

* `network_id` - (Required) Unique identifier of the parent network. Changing this creates a new subnet.

* `cidr` - (Optional) CIDR of the subnet. Must be the network address of its own prefix, for example, `192.168.199.0/24` and not `192.168.199.5/24`; any other value is rejected before the API is called. Changing this creates a new subnet.

* `name` - (Optional) Name of the subnet. If you do not specify a name, the API assigns one.

* `description` - (Optional) Description of the subnet.

* `ip_version` - (Optional) Internet protocol version: `4` (default) or `6`. Changing this creates a new subnet.

* `gateway_ip` - (Optional) Gateway of the subnet. Conflicts with `no_gateway`. If you specify neither, the API assigns a gateway.

* `no_gateway` - (Optional) Creates the subnet without a gateway. The default value is `false`. Conflicts with `gateway_ip`.

* `enable_dhcp` - (Optional) Specifies whether DHCP is enabled for the subnet. The default value is `true`.

* `dns_nameservers` - (Optional) DNS name servers of the subnet, in order. Duplicate entries are rejected before the API is called.

* `allocation_pool` - (Optional) Address ranges available for allocation. You can specify several pools:

  * `start` - (Required) First address of the pool.

  * `end` - (Required) Last address of the pool.

* `ipv6_address_mode` - (Optional) IPv6 address assignment mode: `slaac`, `dhcpv6-stateful`, or `dhcpv6-stateless`. Requires `ip_version` to be `6`. Changing this creates a new subnet.

* `ipv6_ra_mode` - (Optional) IPv6 router advertisement mode, with the same values as `ipv6_address_mode`. Requires `ip_version` to be `6`. Changing this creates a new subnet.

* `tags` - (Optional) Tags of the subnet that Terraform manages. Tags added outside Terraform are reported in `all_tags` and are not removed by this resource.

## Attributes Reference

* `all_tags` - All tags of the subnet, including the ones added outside Terraform.

* `service_types` - Service types associated with the subnet. Read-only.

* `blocked` - Specifies whether the subnet is administratively blocked. Read-only.

~> **Note:** To read the full list of routes of a subnet, including routes added by other resources, use the [selectel_networking_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/networking_subnet_v2) data source, which exports `host_routes`.

## Import

You can import a subnet:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_networking_subnet_v2.subnet_1 <subnet_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).

* `<username>` — Name of the service user.

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. Taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<pool>` — Pool where the subnet is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<subnet_id>` — Unique identifier of the subnet, for example, `8f6c1e9b-2a3d-4f50-9b8a-1c2d3e4f5a6b`.
