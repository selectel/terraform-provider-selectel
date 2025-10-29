---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_os_v1"
sidebar_current: "docs-selectel-datasource-dedicated-os-v1"
description: |-
  Provides a list of available operating systems.
---

# selectel\_dedicated\_os\_v1

Provides a list of available operating systems.

## Example Usage

```hcl
  filter {
    name             = "Ubuntu"
    version_value          = "22.04"
    # version_name     = "22.04 LTS"
    configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
    location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `filter` - (Optional) Values to filter available operating systems.

    * `name` - (Optional) Name of the OS to search.
    * `version_value` - (Optional) Version value of the OS to search. For more information on available OS versions, see the [List OS configurations](https://docs.selectel.ru/en/api/dedicated/#tag/Boot-Manager/operation/get_os_template_list_new_view) method in the Dedicated servers API.
    * `version_name` - (Optional) Version name of the OS to search. Can be a part of name, the search is case-insensitive. For more information on available OS versions, see the [List OS configurations](https://docs.selectel.ru/en/api/dedicated/#tag/Boot-Manager/operation/get_os_template_list_new_view) method in the Dedicated servers API.
    * `configuration_id` - (Optional) Unique identifier of the server configuration. Retrieved from the [selectel_dedicated_configuration_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data_source/dedicated_configuration_v1)
    * `location_id` - (Optional) Unique identifier of the location. Retrieved from the [selectel_dedicated_location_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data_source/dedicated_location_v1). Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/availability-matrix/#dedicated-servers).

## Attributes Reference

* `os` - List of the available operating systems:

    * `id` - Unique identifier of the OS.
    * `name` - OS name.
    * `arch` - OS architecture.
    * `os` - OS type.
    * `version_value` - OS version value raw.
    * `version_name` - OS version name.
    * `scripts_allowed` - Shows if user script is allowed.
    * `ssh_key_allowed` - Shows if SSH key is allowed.
    * `partitioning` - Shows if partitioning is allowed.
