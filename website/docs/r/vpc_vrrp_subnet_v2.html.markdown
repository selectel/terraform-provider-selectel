---
layout: "selectel"
page_title: "Selectel: selectel_vpc_vrrp_subnet_v2"
sidebar_current: "docs-selectel-resource-vpc-vrrp-subnet-v2"
description: |-
  Manages a V2 VRRP subnet resource within Selectel VPC.
---

# selectel\_vpc\_vrrp_subnet_v2

> **WARNING**: this resource has been removed because Selectel VPC Resell V2 API deprecated usage of VRRP subnets.

Manages a V2 VRRP subnet resource within Selectel VPC.

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
}

resource "selectel_vpc_vrrp_subnet_v2" "vrrp_subnet_1" {
  project_id    = "${selectel_vpc_project_v2.project_1.id}"
  master_region = "ru-1"
  slave_region  = "ru-2"
  ip_version    = "ipv4"
  prefix_length = 29
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) An associated Selectel VPC project. Changing this
  creates a new VRRP subnet.

* `master_region` - (Required) A master region of where the VRRP subnet resides.
  Changing this creates a new VRRP subnet.

* `slave_region` - (Required) A slave region of where the VRRP subnet resides.
  Changing this creates a new VRRP subnet.

* `prefix_length` - (Optional) A prefix length of the VRRP subnet. Defaults to 29.
  Changing this creates a new VRRP subnet.

* `ip_version` - (Optional) A version of the IP protocol of the VRRP subnet.
  Defaults to "ipv4". Changing this creates a new VRRP subnet.

## Attributes Reference

The following attributes are exported:

* `cidr` - Shows VRRP subnet CIDR representation.

* `subnets` - Shows information about OpenStack Networking subnets that use this
  VRRP subnet. Contains `network_id`, `subnet_id` and `region` fields.

* `status` - Shows if the VRRP subnet is used or not.

* `servers` - Shows information about servers that use this VRRP subnet. Contains
  `id`, `name` and `status` fields.

## Import

VRRP subnets can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selectel_vpc_vrrp_subnet_v2.vrrp_subnet_1 2060
```
