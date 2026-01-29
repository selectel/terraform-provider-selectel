---
layout: "selectel"
page_title: "Selectel: selectel_global_router_service_v1"
sidebar_current: "docs-selectel-datasource-global-router-service-v1"
description: |-
  Provides a list of services in the Global Router service using public API v1.
---

# selectel\_global\_router\_service\_v1

Provides a list of services in the Global Router service using public API v1.
A service represents a scope of products and services using the same network infrastructure.
For example, the `vpc` service represents cloud servers, file storage, Managed Kubernetes, and Managed Databases.
For more information about global routers, see the [official Selectel documentation](https://docs.selectel.ru/en/global-router/).

## Example Usage

```hcl
data "selectel_global_router_service_v1" "service_1" {
  name = "vpc"
}
```

## Argument Reference

* `name` - (Optional) Service name. Available names are `dedicated`, `vpc`, `vmware`, and `infra`. If the service name is not specified, the data source will return the full list of services.

## Attributes Reference

* `id` - Unique identifier of the service.
* `name` - Service name.
* `extension` - Extension which the Global Router service uses to work with the service. Usually matches the service name.
* `created_at` - Time when the service was created.
