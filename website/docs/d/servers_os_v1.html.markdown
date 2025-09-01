---
layout: "selectel"
page_title: "Selectel: selectel_servers_os_v1"
sidebar_current: "docs-selectel-datasource-servers-os-v1"
description: |-
  Provides a list of available operating systems.
---

# selectel\_servers\_os\_v1

Provides a list of available operating systems.

## Example Usage

```hcl
data "selectel_servers_configuration_v1" "server_configs" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name           = "CL25-NVMe"
    is_server_chip = true
  }
}

data "selectel_servers_location_v1" "server_locations" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name = "MSK-2"
  }
}

data "selectel_servers_os_v1" "server_os" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name             = "Ubuntu"
    version          = "22.04"
    configuration_id = data.selectel_servers_configuration_v1.server_configs.configurations[0].id
    location_id      = data.selectel_servers_location_v1.server_locations.locations[0].id
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `filter` - (Optional) Values to filter available operating systems.

    * `name` - (Optional) Name of the OS to search.
    * `version` - (Optional) Version of the OS to search.
    * `configuration_id` - (Optional) Unique identifier of the server configuration.
    * `location_id` - (Optional) Unique identifier of the location.

## Attributes Reference

* `os` - List of the available operating systems:

    * `id` - Unique identifier of the OS.
    * `name` - OS name.
    * `arch` - OS architecture.
    * `os` - OS type.
    * `version` - OS version.
    * `scripts_allowed` - Shows if user script is allowed.
    * `ssh_key_allowed` - Shows if SSH key is allowed.
    * `partitioning` - Shows if partitioning is allowed.

