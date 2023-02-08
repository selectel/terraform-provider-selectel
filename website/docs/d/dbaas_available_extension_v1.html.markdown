---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_available_extension_v1"
sidebar_current: "docs-selectel-datasource-dbaas-available-extension-v1"
description: |-
  Get information on Selectel DBaaS available extensions.
---

# selectel\_dbaas\_available_extension_v1

Use this data source to get all available extensions within Selectel DBaaS API Service

## Example Usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
}

data "selectel_dbaas_available_extension_v1" "ae" {
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
  filter {
    name  = "hstore"
  }
}
```

## Argument Reference

The folowing arguments are supported

* `project_id` - (Required) An associated Selectel VPC project.

* `region` - (Required) A Selectel VPC region.

* `filter` - (Optional) One or more values used to look up available extensions

**filter**

- `name` - (Optional) Name of the available extension to lookup.

## Attributes Reference

The following attributes are exported:

* `available_extensions` - Contains a list of the found available extensions.

**available_extensions**

- `id` - ID of the extension.
- `name` - Name of the extension.
- `datastore_type_ids` - List of datastore types that support this extension.
- `dependency_ids` - List of extensions that depend on this extension.
