---
layout: "selectel"
page_title: "Selectel: selectel_compute_volume_v3"
sidebar_current: "docs-selectel-datasource-compute-volume-v3"
description: |-
  Provides information about a network volume in Selectel Cloud Servers using public OpenStack Block Storage API v3.
---

# selectel\_compute\_volume_v3

Provides information about an existing network volume in Selectel Cloud Servers using public OpenStack Block Storage API v3. For more information about network volumes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/servers/volumes/about-network-volumes/).

## Example Usage

```hcl
data "selectel_compute_volume_v3" "volume_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-1"
  volume_id  = selectel_compute_volume_v3.volume_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the volume is located, for example, `ru-1`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `volume_id` - (Optional) Unique identifier of the volume. Uses a direct lookup and conflicts with `name`, `status`, and `metadata`. If omitted, the data source applies the configured criteria to the complete volume list and requires exactly one match. With no criteria, the project and pool must contain exactly one visible volume.

* `name` - (Optional) Exact volume name to match.

* `status` - (Optional) Exact volume status to match, for example, `available`.

* `metadata` - (Optional) Key-value pairs that the volume metadata must contain. Other metadata keys do not prevent a match.

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
