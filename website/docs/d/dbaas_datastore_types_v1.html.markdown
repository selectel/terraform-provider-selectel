---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_datastore_types_v1"
sidebar_current: "docs-selectel-datasource-dbaas-datastore-types-v1"
description: |-
  Get information on Selectel DBaaS datastore types.
---

# selectel\_dbaas\_datastore_type_v1

Use this data source to get all available datastore types within Selectel DBaaS API Service

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
}

data "selectel_dbaas_datastore_types_v1" "dt" {
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
  filter       = {
    engine  = "postgresql"
    version = "12"
  }
}
```

## Argument Reference

The folowing arguments are supported

* `project_id` - (Required) An associated Selectel VPC project.

* `region` - (Required) A Selectel VPC region.

* `filter` - (Optional) One or more values used to look up datastore types

**filter**

- `engine` - (Optional) Engine of the datastore type to lookup.
- `version` - (Optional) Version of the datastore type to lookup.

## Attributes Reference

The following attributes are exported:

* `datastore_types` - Contains a list of the found datastore types.

**datastore_types**

- `id` - ID of the datastore type.
- `engine` - Engine of the datastore type.
- `version` - Version of the datastore type.
