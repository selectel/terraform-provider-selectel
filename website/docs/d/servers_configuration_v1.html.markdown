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
data "selectel_servers_configuration_v1" "server_config" {
  project_id = selectel_vpc_project_v2.project_1.id
  deep_filter = file("filter.json")
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `deep_filter` - (Optional) JSON filter. You can use it to filter the results. See an example of the filter values in the [API documentation](https://docs.selectel.ru/en/api/dedicated/#tag/Services/operation/get_server_list).
For example, to get configurations with 1 GPU, with the "Active" state, and that are not manually erasable, you can use the following filter:
```json
{
  "gpu": {
    "count": 1
  },
  "state": "Active",
  "is_manual_erase": false
}
```
Note: Arrays filter checks inclusion, not the full equality.

## Attributes Reference

* `configurations` - List of the available configurations:

  * `id` - Unique identifier of the configuration.

  * `name` - Configuration name.

