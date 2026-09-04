---
layout: "selectel"
page_title: "Selectel: selectel_compute_volume_v3"
sidebar_current: "docs-selectel-resource-compute-volume-v3"
description: |-
  Creates and manages a network volume in Selectel Cloud Servers using public OpenStack Block Storage API v3.
---

# selectel\_compute\_volume_v3

Creates and manages a network volume in Selectel Cloud Servers using public OpenStack Block Storage API v3.
For more information about network volumes, see the [official Selectel documentation](https://docs.selectel.ru/cloud-servers/volumes/about-network-volumes/).

## Example Usage

### Create an empty network volume

```hcl
resource "selectel_compute_volume_v3" "volume_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-1"
  size       = 10
}
```

### Create an SSD Fast v2 network volume with custom IOPS

```hcl
resource "selectel_compute_volume_v3" "volume_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-6"
  size              = 10
  availability_zone = "ru-6a"
  volume_type       = "fast2.ru-6a"

  metadata = {
    total_iops_sec = "30000"
  }
}
```

### Create a network volume from a snapshot

```hcl
resource "selectel_compute_volume_v3" "volume_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-1"
  size        = 20
  snapshot_id = data.selectel_compute_snapshot_v3.snapshot_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.
  Changing this creates a new network volume.
  Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource.
  Learn more about [Projects](https://docs.selectel.ru/access-control/projects/about-projects/).

* `region` - (Required) Pool where the network volume is located, for example, `ru-1`.
  Changing this creates a new network volume.
  Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/infrastructure/availability-matrix/#network-volumes).

* `size` - (Required) Network volume size in GB.
  Must be at least `1`.
  When `snapshot_id`, `source_vol_id`, `image_id`, or `backup_id` is set, the size must be equal to or greater than the source size.
  The maximum size depends on the network volume type, pool segment, and whether the network volume is used as a boot or additional volume.
  Learn more about maximum sizes in the [Network volume limits](https://docs.selectel.ru/cloud-servers/volumes/about-network-volumes/#network-volume-limits).
  You can only increase the size.
  To increase the size of an attached network volume, set `enable_online_resize` to `true`.
  Learn more about [increasing a network volume](https://docs.selectel.ru/cloud-servers/volumes/edit-volume/#enlarge-volume).

* `enable_online_resize` - (Optional) Specifies if Terraform can increase the size of an attached network volume.
  Boolean flag:

  * `false` (default) — Terraform returns an error when you increase the size of an attached network volume.
    Detach the network volume before retrying.
  * `true` — Terraform requests an online resize while the network volume remains attached.

  Learn more about [increasing a network volume](https://docs.selectel.ru/cloud-servers/volumes/edit-volume/#enlarge-volume).

* `name` - (Optional) Network volume name.

* `description` - (Optional) Network volume description.

* `availability_zone` - (Optional) Pool segment where the network volume is located, for example, `ru-1a`.
  Changing this creates a new network volume.
  If omitted, the Block Storage API selects a pool segment.
  Learn more about available pool segments in the [Availability matrix](https://docs.selectel.ru/infrastructure/availability-matrix/#network-volumes).

* `volume_type` - (Optional) Network volume type, for example, `fast.ru-1a`.
  Changing this creates a new network volume.
  Available type prefixes are `basic`, `basicssd`, `universal`, `universal2`, `fast`, and `fast2`.
  Specify the type in the `<volume_type>.<pool_segment>` format.
  We recommend specifying the type that matches `availability_zone`.
  If the API replaces a pool suffix with the network volume's pool segment, the provider treats the returned name as the same type.
  If omitted, the Block Storage API selects the default type.
  This is the project default when configured, otherwise the platform default.
  To retrieve the current default type, set `volume_type_id = "default"` in the [selectel_compute_volume_type_v3](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/compute_volume_type_v3) data source.
  Learn more about available types in the [Network volume types list](https://docs.selectel.ru/cloud-servers/volumes/about-network-volumes/#network-volume-types-list).

* `metadata` - (Optional) Key-value pairs associated with the network volume.
  Metadata keys and values must contain at most 255 characters each.
  The following metadata keys have special behavior:

  * `total_iops_sec` - Configures total read and write IOPS for a Universal v2 or SSD Fast v2 network volume.
    Available values are from `2000` to `16000` for Universal v2 and from `25000` to `75000` for SSD Fast v2.
    The default values are `2000` and `25000`, respectively.
    The Block Storage API does not allow removing this key after network volume creation.
    If you omit an existing value, the provider preserves it.
    Set another supported value to change it.
  * `total_bytes_sec` - Managed by the Block Storage API.
    After the API returns this key, the provider preserves its value and ignores attempts to change or remove it from the configuration.
  * `selectel_tf_create_token` - Reserved by the provider and cannot be configured.

  The API can add other service metadata, which the provider publishes in the state.
  Learn more about available values in the [Network volume limits](https://docs.selectel.ru/cloud-servers/volumes/about-network-volumes/#network-volume-limits).

* `snapshot_id` - (Optional) Unique identifier of a snapshot to use as the network volume source.
  Changing this creates a new network volume.
  Conflicts with `source_vol_id`, `image_id`, and `backup_id`.
  You can retrieve the ID with the [selectel_compute_snapshot_v3](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/compute_snapshot_v3) data source.
  Learn more about [snapshots](https://docs.selectel.ru/cloud-servers/volumes/snapshots/).

* `source_vol_id` - (Optional) Unique identifier of another network volume to copy.
  Changing this creates a new network volume.
  Conflicts with `snapshot_id`, `image_id`, and `backup_id`.
  You can retrieve the ID from the [selectel_compute_volume_v3 resource](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/compute_volume_v3) or [selectel_compute_volume_v3 data source](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/compute_volume_v3).

* `image_id` - (Optional) Unique identifier of an image to use as the network volume source.
  Changing this creates a new network volume.
  Conflicts with `snapshot_id`, `source_vol_id`, and `backup_id`.
  You can retrieve the ID with the [openstack_images_image_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/images_image_v2) data source.

* `backup_id` - (Optional) Unique identifier of a backup to restore.
  Changing this creates a new network volume.
  Conflicts with `snapshot_id`, `source_vol_id`, and `image_id`.
  You can retrieve the ID from the `checkpoints[*].list[*].checkpoint_items[*].backup_id` attribute of the [selectel_cloudbackup_checkpoint_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/cloudbackup_checkpoint_v2) data source.
  Learn more about [backups](https://docs.selectel.ru/cloud-servers/backups/about-backups/).

* `timeouts` - (Optional) How long Terraform waits for volume operations:

  * `create` - (Optional) Timeout for creating a network volume.
    The default value is `10m`.

  * `update` - (Optional) Timeout for waiting for a network volume resize to complete.
    The default value is `20m`.

  * `delete` - (Optional) Timeout for deleting a network volume.
    The default value is `10m`.

## Attributes Reference

* `attachment` - Network volume attachments.
  This attribute is read-only.
  The provider does not manage attachments.
  [Attach or detach the network volume](https://docs.selectel.ru/cloud-servers/volumes/attach-detach-volume/) in the Control panel or with OpenStack CLI.
  Detach the network volume before destroying it.
  If the network volume is attached, Terraform returns an error and keeps it in the state.

  * `id` - Unique identifier of the network volume.
    The Block Storage API repeats the network volume ID in each attachment record; this is not an attachment ID.

  * `instance_id` - Unique identifier of the cloud server.

  * `device` - Device name on the cloud server.

## Import

The `snapshot_id`, `source_vol_id`, `image_id`, and `backup_id` creation sources are not restored during import.
Omit these arguments from the configuration of an imported network volume unless you intend to replace it: adding any of them causes Terraform to create a new network volume and delete the imported one.
The `enable_online_resize` setting is local to Terraform and defaults to `false` after import.

Set the project and pool explicitly because they cannot be derived from the network volume ID.

You can import a volume:

```shell
export OS_AUTH_URL=<auth_url>
export OS_REGION_NAME=<authentication_pool>
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<volume_pool>

terraform import selectel_compute_volume_v3.volume_1 <volume_id>
```

where:

* `<auth_url>` — URL for API authorization.
  See the [list of API URLs](https://docs.selectel.ru/api/urls/).

* `<authentication_pool>` — Pool used for API requests during authorization, for example, `ru-1`.
  This value does not determine the network volume pool.
  Learn more about [API authorization](https://docs.selectel.ru/api/authorization/).

* `<account_id>` — Selectel account ID.
  The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).
  Learn more about [Registration](https://docs.selectel.ru/account/registration/).

* `<username>` — Name of the service user.
  To get the name, in the [Control panel](https://my.selectel.ru/iam/service-users), go to **IAM** ⟶ **Service users** ⟶ copy the name of the required user.
  Learn more about [Service users](https://docs.selectel.ru/access-control/user-types/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project.
  To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Projects** ⟶ copy the ID in the row of the required project.
  Learn more about [Projects](https://docs.selectel.ru/access-control/projects/about-projects/).

* `<volume_pool>` — Pool where the network volume is located, for example, `ru-1`.
  Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/infrastructure/availability-matrix/#network-volumes).

* `<volume_id>` — Unique identifier of the network volume.
  To get the network volume ID, use the `openstack --os-region-name <volume_pool> volume list` command.
  Learn more about [OpenStack CLI](https://docs.selectel.ru/cloud-servers/tools/openstack-cli/).
