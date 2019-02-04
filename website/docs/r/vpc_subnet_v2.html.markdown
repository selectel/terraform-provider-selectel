---
layout: "selectel"
page_title: "Selectel: selectel_vpc_subnet_v2"
sidebar_current: "docs-selectel-resource-vpc-subnet-v2"
description: |-
  Manages a V2 subnet resource within Selectel VPC.
---

# selectel\_vpc\_subnet_v2

Manages a V2 subnet resource within Selectel VPC.

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
}

resource "selectel_vpc_subnet_v2" "subnet_1" {
  project_id    = "${selectel_vpc_project_v2.project_1.id}"
  region        = "ru-3"
  ip_version    = "ipv4"
  prefix_length = 29
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) An associated Selectel VPC project. Changing this
  creates a new subnet.

* `region` - (Required) A region of where the subnet resides. Changing this
  creates a new subnet.

* `prefix_length` - (Optional) A prefix length of the subnet. Defaults to 29.
  Changing this creates a new subnet.

* `ip_version` - (Optional) A version of the IP protocol of the subnet. Defaults
  to "ipv4". Changing this creates a new subnet.

## Attributes Reference

The following attributes are exported:

* `cidr` - Shows subnet CIDR representation.

* `network_id` - Shows associated OpenStack Networking service network ID.

* `subnet_id` - Shows associated OpenStack Networking service subnet ID.

* `status` - Shows if the subnet is used or not.

* `servers` - Shows information about servers that use this subnet. Contains
  `id`, `name` and `status` fields.

## Import

Subnets can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selectel_vpc_subnet_v2.subnet_1 2060
```
