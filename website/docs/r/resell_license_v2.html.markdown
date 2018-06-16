---
layout: "selvpc"
page_title: "SelVPC: selvpc_resell_license_v2"
sidebar_current: "docs-selvpc-resource-resell-license-v2"
description: |-
  Manages a V2 license resource within Resell Selectel VPC.
---

# selvpc\_resell\_license_v2

Manages a V2 license resource within Resell Selectel VPC.

## Example Usage

```hcl
resource "selvpc_resell_license_v2" "license_windows_2016_standard" {
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

* `type` - (Required) A type of the license. Changing this creates a new license.

## Attributes Reference

The following attributes are exported:

* `project_id` - See Argument Reference above.

* `region` - See Argument Reference above.

* `status` - Shows if the license is used or not.

* `servers` - Shows information about servers that use this license. Contains
  `id`, `name` and `status` fields.

* `type` - See Argument Reference above.

## Import

Licenses can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selvpc_resell_license_v2.license_1 4123
```
