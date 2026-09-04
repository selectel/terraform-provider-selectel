---
layout: "selectel"
page_title: "Selectel: selectel_compute_volume_v3"
sidebar_current: "docs-selectel-datasource-compute-volume-v3"
description: |-
  Provides a network volume in Selectel Cloud Servers.
---

# selectel\_compute\_volume_v3

Provides a network volume in Selectel Cloud Servers. For more information about network volumes, see the [official Selectel documentation](https://docs.selectel.ru/cloud-servers/volumes/about-network-volumes/).

## Example Usage

```hcl
data "selectel_compute_volume_v3" "volume_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-1"
  volume_id  = selectel_compute_volume_v3.volume_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/access-control/projects/about-projects/).

* `region` - (Required) Pool where the volume is located, for example, `ru-1`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/infrastructure/availability-matrix/#network-volumes).

* `volume_id` - (Optional) Unique identifier of the volume. Uses a direct lookup and conflicts with `name`, `status`, and `metadata`. If omitted, the data source applies the configured criteria to the complete volume list and requires exactly one match. With no criteria, the project and pool must contain exactly one visible volume. Retrieved from the [selectel_compute_volume_v3](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/compute_volume_v3) resource.

* `name` - (Optional) Exact volume name to match. Conflicts with `volume_id`.

* `status` - (Optional) Exact volume status to match, for example, `available`. Conflicts with `volume_id`.

* `metadata` - (Optional) Key-value pairs that the volume metadata must contain. Other metadata keys do not prevent a match. Conflicts with `volume_id`. The key `selectel_tf_create_token` is reserved by the provider and cannot be used as a search criterion.

## Attributes Reference

* `id` - Unique identifier of the found volume.

* `name` - Volume name.

* `description` - Volume description.

* `size` - Volume size in GB.

* `status` - Volume status.

* `availability_zone` - Pool segment where the volume is located.

* `volume_type` - Volume type returned by the Block Storage API.

* `bootable` - Whether the volume is bootable. The Block Storage API returns this value as a string.

* `metadata` - Volume metadata returned by the Block Storage API.

* `snapshot_id` - Unique identifier of the source snapshot, if the volume was created from a snapshot.

* `source_vol_id` - Unique identifier of the source volume, if the volume was copied from another volume.

* `attachment` - Volume attachments observed through the Block Storage API.

  * `id` - Unique identifier of the volume. The Block Storage API repeats the volume ID in each attachment record; this is not an attachment ID.

  * `instance_id` - Unique identifier of the cloud server.

  * `device` - Device name on the cloud server.
