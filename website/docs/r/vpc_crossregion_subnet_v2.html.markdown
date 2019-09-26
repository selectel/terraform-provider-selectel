---
layout: "selectel"
page_title: "Selectel: selectel_vpc_crossregion_subnet_v2"
sidebar_current: "docs-selectel-resource-vpc-crossregion-subnet-v2"
description: |-
  Manages a V2 Cross-region subnet resource within VPC Selectel VPC.
---

# selectel\_vpc\_crossregion_subnet_v2

> **WARNING**: this resource has been removed because Selectel VPC Resell V2 API deprecated usage of crossregion subnets.

Manages a V2 Cross-region subnet resource within VPC Selectel VPC.

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
}

resource "selectel_vpc_crossregion_subnet_v2" "crossregion_subnet_1" {
  project_id = "${selectel_vpc_project_v2.project_1.id}"
  cidr = "192.168.200.0/24"
  regions {
    region = "ru-1"
  }
  regions {
    region = "ru-3"
  }
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) An associated Selectel VPC project. Changing this
  creates a new Cross-region subnet.

* `regions` - (Required) An array of regions where the Cross-region subnet resides.
  Changing this creates a new Cross-region subnet. The structure is described below.

* `cidr` - (Required) A cross-region subnet CIDR representation. Changing this
  creates a new Cross-region subnet.

The `regions` block supports:

* `region` - (Required) A region of where the Cross-region subnet resides.
  Changing this creates a new Cross-region subnet.

## Attributes Reference

The following attributes are exported:

* `servers` - Shows information about servers that use this Cross-region subnet. Contains
  `id`, `name` and `status` fields.

* `status` - Shows if the Cross-region subnet is used or not.

* `subnets` - Shows information about OpenStack Networking subnets that use this
  Cross-region subnet. Contains `cidr`, `network_id`, `project_id`, `region`, `subnet_id`,
  `vlan_id` and `vtep_ip_address` fields.

* `vlan_id` - Shows id of the associated VLAN in the OpenStack Networking service for
  this Cross-region subnet.

## Import

Cross-region subnets can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selectel_vpc_crossregion_subnet_v2.crossregion_subnet_1 2060
```