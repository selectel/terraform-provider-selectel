---
layout: "selectel"
page_title: "Selectel: selectel_vpc_license_v2"
sidebar_current: "docs-selectel-resource-vpc-license-v2"
description: |-
  Manages a V2 license resource within Selectel VPC.
---

# selectel\_vpc\_license_v2

Manages a V2 license resource within Selectel VPC.

## Example Usage

```hcl
resource "selectel_vpc_license_v2" "license_windows_2016_standard" {
  project_id = "887e5e35458d4ee38a6ae0543555dec5"
  region     = "ru-2"
  type       = "license_windows_2012_standard"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) An associated Selectel VPC project. Changing this
  creates a new license.

* `region` - (Required) A region of where the license resides. Changing this
  creates a new license.

* `type` - (Required) The type of license. Changing this creates a new license.

## Attributes Reference

The following attributes are exported:

* `status` - Shows if the license is used or not.

* `servers` - Shows information about servers that use this license. Contains
  `id`, `name` and `status` fields.

* `network_id` - Represents id of the associated network in the Networking service.

* `subnet_id` - Represents id of the associated network in the Networking service.

## Import

Licenses can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selectel_vpc_license_v2.license_1 4123
```
