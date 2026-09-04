---
layout: "selectel"
page_title: "Selectel: selectel_compute_volume_type_v3"
sidebar_current: "docs-selectel-datasource-compute-volume-type-v3"
description: |-
  Provides a network volume type in Selectel Cloud Servers.
---

# selectel\_compute\_volume\_type_v3

Provides a network volume type in Selectel Cloud Servers. For more information about volume types, see the [official Selectel documentation](https://docs.selectel.ru/cloud-servers/volumes/about-network-volumes/).

## Example Usage

```hcl
data "selectel_compute_volume_type_v3" "volume_type_1" {
  project_id     = selectel_vpc_project_v2.project_1.id
  region         = "ru-1"
  volume_type_id = "default"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/access-control/projects/about-projects/).

* `region` - (Required) Pool where the volume type is available, for example, `ru-1`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/infrastructure/availability-matrix/#network-volumes).

* `volume_type_id` - (Optional) Unique identifier of the volume type or the special value `default`. A successful lookup stores the actual UUID in the data source `id`. Conflicts with `name`. Exactly one of `volume_type_id` or `name` is required. See available IDs in the [Network volume types list](https://docs.selectel.ru/cloud-servers/volumes/about-network-volumes/#network-volume-types-list).

* `name` - (Optional) Exact volume type name. Available type prefixes are `basic`, `basicssd`, `universal`, `universal2`, `fast`, and `fast2`. The zonal type format is `<volume_type>.<pool_segment>`. The data source reads the complete public type list and requires exactly one match. Conflicts with `volume_type_id`. Exactly one of `volume_type_id` or `name` is required. See available names in the [Network volume types list](https://docs.selectel.ru/cloud-servers/volumes/about-network-volumes/#network-volume-types-list).

## Attributes Reference

* `id` - Actual unique identifier of the found volume type. When `volume_type_id = "default"`, this is the UUID resolved by the API rather than the string `default`.

* `name` - Volume type name returned by the API.

* `description` - Volume type description.

* `is_public` - Whether the volume type is public.

* `extra_specs` - Extra specifications visible to the current role. Administrative specifications that the API does not return are not added to the state.

* `supports_custom_iops` - Whether the volume type supports user-managed IOPS through `metadata.total_iops_sec` on a volume.
