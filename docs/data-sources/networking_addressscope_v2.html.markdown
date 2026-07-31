---
layout: "selectel"
page_title: "Selectel: selectel_networking_addressscope_v2"
sidebar_current: "docs-selectel-datasource-networking-addressscope-v2"
description: |-
  Provides information about an address scope in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_addressscope\_v2

Provides information about an address scope.

The data source returns exactly one address scope. If no address scope matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

## Example Usage

```hcl
data "selectel_networking_addressscope_v2" "addressscope_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "addressscope-1"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.

* `region` - (Required) Selectel region where the address scope is located, for example, `ru-3`.

You can combine any of the following selectors. The address scope must match all of them at the same time:

* `name` - (Optional, Computed) Name of the address scope.

* `ip_version` - (Optional, Computed) Internet protocol version of the address scope, `4` or `6`.

* `shared` - (Optional, Computed) Specifies whether the address scope is shared with other projects. Selectable here even though the resource cannot set it, so you can look up the scopes an administrator shared with you.

## Attributes Reference

Every attribute above is filled from the address scope that was found, not from the value you passed as a selector. The identifier of the data source is the identifier of that address scope.

~> **Note:** The returned attributes report the actual state of the address scope and cannot be used to change it.
