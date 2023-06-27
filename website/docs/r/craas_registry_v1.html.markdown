---
layout: "selectel"
page_title: "Selectel: selectel_craas_registry_v1"
sidebar_current: "docs-selectel-resource-craas-registry-v1"
description: |-
Manages a V1 registry resource within Selectel Container Registry Service.
---

# selectel\_craas\_registry\_v1

Manages a V1 registry resource within Selectel Container Registry Service.

## Example usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  name = "my-first-project"
}

resource "selectel_craas_registry_v1" "registry_1" {
  name       = "my-first-registry"
  project_id = selectel_vpc_project_v2.project_1.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the registry.
  Changing this creates a new registry.

* `project_id` - (Required) An associated Selectel VPC project.
  Changing this creates a new registry.

## Attributes Reference

The following attributes are exported:

* `status` - Shows the current status of the registry.

## Import

Registry can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN SEL_PROJECT_ID=SELECTEL_VPC_PROJECT_ID terraform import selectel_craas_registry_v1.registry_1 939506d6-7621-4581-b673-eacf3db30f5b
```
