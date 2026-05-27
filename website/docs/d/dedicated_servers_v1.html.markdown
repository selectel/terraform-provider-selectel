---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_servers_v1"
sidebar_current: "docs-selectel-datasource-dedicated-servers-v1"
description: |-
  Provides a list of dedicated servers in the project.
---

# selectel\_dedicated\_servers\_v1

Provides a list of dedicated servers in the project. Learn more about [About dedicated servers.](https://docs.selectel.ru/en/dedicated/about/about-dedicated)

## Example Usage

### Get all servers in project

```hcl
data "selectel_dedicated_servers_v1" "servers" {
  project_id = selectel_vpc_project_v2.project.id
  filter {
    name = "production-web-01"
    ip = "192.168.1.100"
    location_id   = data.selectel_location_v1.server_location.locations[0].id
    configuration = "EL5"
    private_subnet = data.selectel_dedicated_private_subnet_v1.subnets.subnets[0].id
    public_subnet = data.selectel_dedicated_public_subnet_v1.subnets.subnets[0].id
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `filter` - (Optional) Values to filter available servers:

  * `name` - (Optional) Name of the server. Supports partial match, case-insensitive.

  * `ip` - (Optional) IP address of the server.

  * `location_id` - (Optional) Unique identifier of the location. Retrieved from the [selectel_dedicated_location_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dedicated_location_v1) data source. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/availability-matrix/#dedicated-servers).

  * `configuration` - (Optional) Partial, case-insensitive substring of the configuration display name (e.g. `"EL50 SSD"`). Matches any server whose configuration name contains the specified string.

  * `public_subnet` - (Optional) Unique identifier of a public subnet to which the server belongs. Retrieved from the [selectel_dedicated_public_subnet_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dedicated_public_subnet_v1). 

  * `private_subnet` - (Optional) Unique identifier of a private subnet to which the server belongs. Retrieved from the [selectel_dedicated_private_subnet_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dedicated_private_subnet_v1).

## Attributes Reference

* `servers` - List of the available servers:

  * `id` - Unique identifier of the server.
  
  * `name` - Server name.
  
  * `configuration_id` - Configuration ID of the server.
  
  * `location_id` - Location ID of the server.
  
  * `reserved_public_ips` - List of reserved public IP addresses for the server.
  
  * `reserved_private_ips` - List of reserved private IP addresses for the server.
