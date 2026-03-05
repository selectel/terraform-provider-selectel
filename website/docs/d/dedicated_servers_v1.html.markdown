---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_servers_v1"
sidebar_current: "docs-selectel-datasource-dedicated-servers-v1"
description: |-
  Provides a list of dedicated servers available in the project.
---

# selectel\_dedicated\_servers\_v1

Provides a list of dedicated servers available in the project.

## Example Usage

### Get all servers in project

```hcl
data "selectel_dedicated_servers_v1" "servers" {
  project_id = selectel_vpc_project_v2.project.id
}
```

### Filter servers by name

```hcl
data "selectel_dedicated_servers_v1" "production_servers" {
  project_id = selectel_vpc_project_v2.project.id
  
  filter {
    name = "production-web-01"
  }
}
```

### Filter servers by IP address

```hcl
data "selectel_dedicated_servers_v1" "server_by_ip" {
  project_id = selectel_vpc_project_v2.project.id
  
  filter {
    ip = "192.168.1.100"
  }
}
```

### Filter servers by location and configuration

```hcl
data "selectel_dedicated_servers_v1" "filtered_servers" {
  project_id = selectel_vpc_project_v2.project.id
  
  filter {
    location_id      = "796f1f0a-d97d-4a8e-904e-4fd5ef57465c"
    configuration_id = "796f1f0a-d97d-4a8e-904e-4fd5ef574652"
  }
}
```

### Filter servers by subnet

```hcl
data "selectel_dedicated_servers_v1" "servers_by_subnet" {
  project_id = selectel_vpc_project_v2.project.id
  
  filter {
    public_subnet = "subnet-public-1"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `filter` - (Optional) Block filter for servers:

  * `name` - (Optional) Name of the server to filter. Supports partial match (case-insensitive).

  * `ip` - (Optional) IP address of the server to filter.

  * `location_id` - (Optional) Unique identifier of the location. Retrieved from the [selectel_dedicated_location_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dedicated_location_v1) data source. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/availability-matrix/#dedicated-servers).

  * `configuration_id` - (Optional) Unique identifier of the server configuration (UUID). Retrieved from the [selectel_dedicated_configuration_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data_source/dedicated_configuration_v1)

  * `public_subnet` - (Optional) Public subnet ID of the server to filter.

  * `private_subnet` - (Optional) Private subnet ID of the server to filter.

## Attributes Reference

* `servers` - List of the available servers:

  * `id` - Unique identifier of the server (UUID).
  
  * `name` - Server name.
  
  * `configuration_id` - Configuration ID of the server.
  
  * `location_id` - Location ID of the server.
  
  * `reserved_public_ips` - List of reserved public IP addresses for the server.
  
  * `reserved_private_ips` - List of reserved private IP addresses for the server.