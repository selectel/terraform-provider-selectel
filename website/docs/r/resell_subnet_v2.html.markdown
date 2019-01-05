---
layout: "selvpc"
page_title: "SelVPC: selvpc_resell_subnet_v2"
sidebar_current: "docs-selvpc-resource-resell-subnet-v2"
description: |-
  Manages a V2 subnet resource within Resell Selectel VPC.
---

# selvpc\_resell\_subnet_v2

Manages a V2 subnet resource within Resell Selectel VPC.

## Example Usage

```hcl
resource "selvpc_resell_project_v2" "project_1" {
  auto_quotas = true
}

resource "selvpc_resell_subnet_v2" "subnet_1" {
  project_id    = "${selvpc_resell_project_v2.project_1.id}"
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
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selvpc_resell_subnet_v2.subnet_1 2060
```
