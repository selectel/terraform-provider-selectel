---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_configuration_v1"
sidebar_current: "docs-selectel-datasource-dedicated-configuration-v1"
description: |-
  Provides a list of server configurations available in Selectel.
---

# selectel\_dedicated\_configuration\_v1

Provides a list of server configurations available in Selectel.

## Example Usage

### Find configuration ID by name

```hcl
data "selectel_dedicated_configuration_v1" "server_config" {
  project_id  = selectel_vpc_project_v2.project_1.id
  deep_filter = "{\"name\":\"CL25-NVMe\"}"
}
```

### Search available configurations with additional parameters


```hcl
data "selectel_dedicated_configuration_v1" "server_config" {
  project_id  = selectel_vpc_project_v2.project_1.id
  deep_filter = <<EOT
    {
        "gpu": {
           "count": 1
        },
        "state": "Active",
    }
  EOT
}
```

### Search available configurations with additional parameters from file


```hcl
data "selectel_dedicated_configuration_v1" "server_config" {
  project_id  = selectel_vpc_project_v2.project_1.id
  deep_filter = file("filter.json")
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `deep_filter` — (Optional) JSON filter for available locations:
  * You can use only the name of the configuration to get the results.To get the name of the configuration, in the [Selectel site](https://selectel.ru/en/services/dedicated/).
  * You can use [additional parameters](#search-available-configurations-with-additional-parameters) or their combinations to filter available configurations. You can set them in place or use another [file](#search-available-configurations-with-additional-parameters-from-file). See an example of the filter values in the [API documentation](https://docs.selectel.ru/en/api/dedicated/#tag/Services/operation/get_server_list)

## Attributes Reference

* `configurations` - List of the available configurations:

  * `id` - Unique identifier of the configuration.

  * `name` - Configuration name.
