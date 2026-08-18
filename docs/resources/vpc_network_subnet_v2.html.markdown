---
layout: "selectel"
page_title: "Selectel: selectel_vpc_network_subnet_v2"
sidebar_current: "docs-selectel-resource-vpc-network-subnet-v2"
description: |-
  Creates and manages a subnet of a private network for Selectel products using public API v2.
---

# selectel\_vpc\_network\_subnet\_v2

Creates and manages a subnet of a private network using public API v2. For more information about private networks and subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/).

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_network_subnet_v2" "subnet_1" {
  project_id      = selectel_vpc_project_v2.project_1.id
  region          = "ru-3"
  network_id      = selectel_vpc_network_v2.network_1.id
  cidr            = "192.168.199.0/24"
  name            = "subnet-1"
  dns_nameservers = ["8.8.8.8", "1.1.1.1"]

  allocation_pool {
    start = "192.168.199.10"
    end   = "192.168.199.100"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new subnet. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource.

* `region` - (Required) Pool where the subnet is located, for example, `ru-3`. Changing this creates a new subnet. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `network_id` - (Required) Unique identifier of the parent network. Changing this creates a new subnet.

* `cidr` - (Required) CIDR of the subnet. Must be the network address of its own prefix, for example, `192.168.199.0/24` and not `192.168.199.5/24`; any other value is rejected before the API is called. Changing this creates a new subnet.

* `name` - (Optional) Name of the subnet. If you do not specify a name, the API assigns one.

* `description` - (Optional) Description of the subnet.

* `ip_version` - (Optional) Internet protocol version: `4` (default) or `6`. Changing this creates a new subnet.

* `gateway_ip` - (Optional) Gateway of the subnet. Conflicts with `no_gateway`. If you specify neither, the API assigns a gateway.

* `no_gateway` - (Optional) Creates the subnet without a gateway. The default value is `false`. Conflicts with `gateway_ip`.

* `enable_dhcp` - (Optional) Specifies whether DHCP is enabled for the subnet. The default value is `true`.

* `dns_nameservers` - (Optional) DNS name servers passed to instances through DHCP, in order. This setting does not create Private DNS records. Duplicate entries are rejected before the API is called.

* `allocation_pool` - (Optional) Address ranges available for allocation. You can specify several pools:

  * `start` - (Required) First address of the pool.

  * `end` - (Required) Last address of the pool.

* `ipv6_address_mode` - (Optional) IPv6 address assignment mode: `slaac`, `dhcpv6-stateful`, or `dhcpv6-stateless`. Requires `ip_version` to be `6`. Changing this creates a new subnet.

* `ipv6_ra_mode` - (Optional) IPv6 router advertisement mode, with the same values as `ipv6_address_mode`. Requires `ip_version` to be `6`. Changing this creates a new subnet.

* `tags` - (Optional) Tags of the subnet that Terraform manages. Tags added outside Terraform are reported in `all_tags` and are not removed by this resource.

## Attributes Reference

* `all_tags` - All tags of the subnet, including the ones added outside Terraform.

* `service_types` - Unordered set of service types associated with the subnet.

* `blocked` - Specifies whether the subnet is administratively blocked.

* `host_routes` - All routes of the subnet, including the ones added by the [selectel_vpc_network_subnet_route_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_route_v2) resource. The subnet resource reports this list but does not manage it:

  * `destination_cidr` - Destination CIDR of the route.

  * `next_hop` - Next hop of the route.

## Import

You can import a subnet:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_network_subnet_v2.subnet_1 <subnet_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). The import ID does not contain the project, so `INFRA_PROJECT_ID` must be set.

* `<pool>` — Pool where the subnet is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/). The import ID does not contain the pool, so `INFRA_REGION` must be set.

* `<subnet_id>` — Unique identifier of the subnet, for example, `8f6c1e9b-2a3d-4f50-9b8a-1c2d3e4f5a6b`. To get the subnet ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ the network page ⟶ the **Subnets** tab ⟶ copy the ID of the required subnet. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack subnet list` and copy the `ID` field.
