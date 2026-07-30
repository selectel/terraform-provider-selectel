---
layout: "selectel"
page_title: "Selectel: selectel_compute_volume_v3"
sidebar_current: "docs-selectel-resource-compute-volume-v3"
description: |-
  Creates and manages a network volume in Selectel Cloud Servers using public OpenStack Block Storage API v3.
---

# selectel\_compute\_volume_v3

Creates and manages a network volume in Selectel Cloud Servers using public OpenStack Block Storage API v3. For more information about network volumes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/servers/volumes/about-network-volumes/).

## Example Usage

### Create an empty volume

```hcl
resource "selectel_compute_volume_v3" "volume_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-1"
  name              = "volume-1"
  description       = "Volume managed by Terraform"
  size              = 10
  availability_zone = "ru-1a"
  volume_type       = "fast.ru-1a"
}
```

### Create a Fast v2 volume with custom IOPS

```hcl
resource "selectel_compute_volume_v3" "volume_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-6"
  name              = "volume-1"
  description       = "Volume managed by Terraform"
  size              = 10
  availability_zone = "ru-6a"
  volume_type       = "fast2.ru-6a"

  metadata = {
    total_iops_sec = "30000"
  }
}
```

### Create a volume from a snapshot

```hcl
resource "selectel_compute_volume_v3" "volume_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-1"
  name        = "volume-from-snapshot"
  size        = 20
  snapshot_id = var.snapshot_id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new volume and deletes the previous one. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the volume is located, for example, `ru-1`. Changing this creates a new volume and deletes the previous one. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `size` - (Required) Volume size in GB. Must be at least `1`. You can only increase the size. To increase the size of an attached volume, set `enable_online_resize` to `true`.

* `enable_online_resize` - (Optional) Enables increasing the size of an attached volume. If set to `false`, detach the volume before increasing its size. Default is `false`.

* `name` - (Optional) Volume name.

* `description` - (Optional) Volume description.

* `availability_zone` - (Optional) Pool segment where the volume is located, for example, `ru-1a`. Changing this creates a new volume and deletes the previous one. If omitted, the Block Storage API selects a pool segment.

* `volume_type` - (Optional) Volume type, for example, `fast.ru-1a`. We recommend specifying the zonal type name that matches `availability_zone`. Changing the type creates a new volume and deletes the previous one. If the API replaces a regional suffix with the volume availability zone, the provider treats the returned name as the same type. If omitted, the Block Storage API selects the default type.

* `metadata` - (Optional) Key-value pairs associated with the volume. For a zonal Fast v2 volume, set `total_iops_sec` to request custom IOPS. Cinder validates whether the selected volume type supports custom IOPS and whether the value is within the current range. The key `selectel_tf_create_token` is reserved by the provider and cannot be configured. The API can add other service metadata, which the provider publishes in the state.

* `snapshot_id` - (Optional) Unique identifier of a snapshot to use as the volume source. Changing this creates a new volume and deletes the previous one. Conflicts with `source_vol_id`, `image_id`, and `backup_id`.

* `source_vol_id` - (Optional) Unique identifier of another volume to copy. Changing this creates a new volume and deletes the previous one. Conflicts with `snapshot_id`, `image_id`, and `backup_id`. You can retrieve the ID from the [selectel_compute_volume_v3](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/compute_volume_v3) resource or data source.

* `image_id` - (Optional) Unique identifier of an image to use as the volume source. Changing this creates a new volume and deletes the previous one. Conflicts with `snapshot_id`, `source_vol_id`, and `backup_id`.

* `backup_id` - (Optional) Unique identifier of a backup to restore. Changing this creates a new volume and deletes the previous one. Conflicts with `snapshot_id`, `source_vol_id`, and `image_id`.

## Attributes Reference

* `attachment` - Volume attachments observed through the Block Storage API. This attribute is read-only; attach and detach the volume through the Cloud Servers API or other supported tools. Detach the volume before destroying it. If the volume is attached, Terraform returns an error and keeps it in the state.

  * `id` - Unique identifier of the volume. The Block Storage API repeats the volume ID in each attachment record; this is not an attachment ID.

  * `instance_id` - Unique identifier of the cloud server.

  * `device` - Device name on the cloud server.

## Import

You can import a volume by its unique identifier. Set the project and pool explicitly because they cannot be derived from the volume ID:

```shell
export OS_AUTH_URL=<keystone_url>
export OS_REGION_NAME=<authentication_pool>
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<volume_pool>

terraform import selectel_compute_volume_v3.volume_1 <volume_id>
```

To get the volume ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Cloud Servers** ⟶ **Volumes**, or use the [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack volume list`.
