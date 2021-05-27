---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_flavor_v1"
sidebar_current: "docs-selectel-datasource-dbaas-flavor-v1"
description: |-
  Get information on Selectel DBaaS flavors.
---

# selectel\_dbaas\_flavors_v1

Use this data source to get all available flavors within Selectel DBaaS API Service

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
}

data "selectel_dbaas_flavor_v1" "flavor" {
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
}
```

## Argument Reference

The folowing arguments are supported

* `project_id` - (Required) An associated Selectel VPC project.

* `region` - (Required) A Selectel VPC region.

## Attributes Reference

The following attributes are exported:

* `flavors` - Contains a list of the found flavors.

**flavors**

- `id` - ID of the flavor.
- `name` - Name of the flavor.
- `description` - Description of the flavor.
- `vcpus` - CPU count for the flavor.
- `ram` - RAM count for the flavor.
- `disk` - Disk size for the flavor.
