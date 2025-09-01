---
layout: "selectel"
page_title: "Selectel: selectel_servers_configuration_v1"
sidebar_current: "docs-selectel-datasource-servers-configuration-v1"
description: |-
  Provides a list of server configurations available in Selectel.
---

# selectel\_servers\_configuration\_v1

Provides a list of server configurations available in Selectel.

## Example Usage

```hcl
data "selectel_servers_configuration_v1" "server_configs" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    is_server_chip = true
    name           = "CL25-NVMe"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `filter` - (Required) Values to filter available configurations.

  * `is_server_chip` - (Required) Specifies the type of server. `true` for Chipcore line, `false` for standard servers.

  * `name` - (Optional) Name of the configuration to search.

## Attributes Reference

* `configurations` - List of the available configurations:

  * `id` - Unique identifier of the configuration.

  * `name` - Configuration name.

