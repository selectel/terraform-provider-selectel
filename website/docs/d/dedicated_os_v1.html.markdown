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
data "selectel_dedicated_configuration_v1" "server_config" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name = "CL25-NVMe"
  }
}

data "selectel_dedicated_location_v1" "server_location" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name = "MSK-2"
  }
}

data "selectel_dedicated_os_v1" "server_os" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name             = "Ubuntu"
    version_value          = "22.04"
    # version_name     = "22.04 LTS"
    # version_name_regex    = "22\\.04"
    configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
    location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `filter` - (Optional) Values to filter available operating systems.

    * `name` - (Optional) Name of the OS to search.
    * `version_value` - (Optional) Version value of the OS to search.
    * `version_name` - (Optional) Version name of the OS to search. Can be a part of name, the search is case-insensitive.
    * `version_name_regex` - (Optional) Version RE2 regex to search OS by version name.
    * `configuration_id` - (Optional) Unique identifier of the server configuration.
    * `location_id` - (Optional) Unique identifier of the location.

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

