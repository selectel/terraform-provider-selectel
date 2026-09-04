---
layout: "selectel"
page_title: "Selectel: selectel_compute_snapshot_v3"
sidebar_current: "docs-selectel-datasource-compute-snapshot-v3"
description: |-
  Provides a network volume snapshot in Selectel Cloud Servers.
---

# selectel\_compute\_snapshot\_v3

Provides a network volume snapshot in Selectel Cloud Servers. For more information about snapshots, see the [official Selectel documentation](https://docs.selectel.ru/cloud-servers/volumes/snapshots/).

The data source is read-only. It does not accept a snapshot ID, cannot be imported, and does not create, update, or delete snapshots.

## Example Usage

```hcl
data "selectel_compute_snapshot_v3" "snapshot_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-1"
  name       = "snapshot-1"
  volume_id  = selectel_compute_volume_v3.volume_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/access-control/projects/about-projects/).

* `region` - (Required) Pool where the snapshot is located, for example, `ru-1`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/infrastructure/availability-matrix/#network-volumes).

* `name` - (Optional) Exact snapshot name to match.

* `status` - (Optional) Exact snapshot status to match, for example, `available`.

* `volume_id` - (Optional) Unique identifier of the source volume. Retrieved from the [selectel_compute_volume_v3](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/compute_volume_v3) resource.

* `most_recent` - (Optional) Specifies if the data source selects a snapshot with the latest `created_at` timestamp when multiple snapshots match. Boolean flag, the default value is `false`. If `false`, the search must return exactly one snapshot. If several matching snapshots have the same latest timestamp, any one of them can be selected.

All string arguments must contain at least one non-whitespace character. You can combine `name`, `status`, and `volume_id`. Only specified filters are sent to the Block Storage API. If you omit all filters, the data source searches all snapshots visible in the project and pool. A search that returns no snapshots fails. A search that returns multiple snapshots fails unless `most_recent` is set to `true`.

## Attributes Reference

* `id` - Unique identifier of the found snapshot.

* `name` - Snapshot name.

* `description` - Snapshot description.

* `volume_id` - Unique identifier of the source volume.

* `status` - Snapshot status.

* `size` - Snapshot size in GB.

* `metadata` - Snapshot metadata returned by the Block Storage API.
