---
layout: "selvpc"
page_title: "SelVPC: selvpc_resell_vrrp_subnet_v2"
sidebar_current: "docs-selvpc-resource-resell-vrrp-subnet-v2"
description: |-
  Manages a V2 VRRP subnet resource within Resell Selectel VPC.
---

# selvpc\_resell\_vrrp_subnet_v2

Manages a V2 VRRP subnet resource within Resell Selectel VPC.

## Example Usage

```hcl
resource "selvpc_resell_project_v2" "project_1" {
  auto_quotas = true
}

resource "selvpc_resell_vrrp_subnet_v2" "vrrp_subnet_1" {
  project_id    = "${selvpc_resell_project_v2.project_1.id}"
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
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selvpc_resell_vrrp_subnet_v2.vrrp_subnet_1 2060
```
