---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_location_v1"
sidebar_current: "docs-selectel-datasource-dedicated-location-v1"
description: |-
  Provides a list of available locations.
---

# selectel\_dedicated\_location\_v1

Provides a list of available locations.

## Example Usage

```hcl
data "selectel_dedicated_location_v1" "server_location" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name = "SPB-2"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `filter` - (Optional) Values to filter available locations.

  * `name` - (Optional) Name of the location to search. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/availability-matrix/#dedicated-servers).

## Attributes Reference

* `locations` - List of the available locations:

  * `id` - Unique identifier of the location.

  * `name` - Location name.

  * `description` - Location description.

  * `visibility` - Location visibility.
