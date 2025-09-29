---
layout: "selectel"
page_title: "Selectel: selectel_cloudbackup_checkpoint_v2"
sidebar_current: "docs-selectel-datasource-cloudbackup-checkpoint-v2"
description: |-
  Provides a list of backup checkpoints for Selectel Scheduled Backup service.
---

# selectel\_cloudbackup\_checkpoint\_v2

Provides a list of backup checkpoints for Selectel Scheduled Backup service.. For more information about checkpoints, see the official Selectel documentation for [Scheduled Backup](https://docs.selectel.ru/en/api/scheduled-backups/).

## Example Usage

```hcl
data "selectel_cloudbackup_checkpoint_v2" "checkpoint_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    plan_name   = "my-backup-plan"
    volume_name = "my-volume"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Region where the backup plan is located, for example, `ru-3`. Learn more about available regions in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `filter` - (Optional) Values to filter checkpoints.

  * `plan_name` - (Optional) Name of the backup plan.
  * `volume_name` - (Optional) Name of the volume.

## Attributes Reference

* `checkpoints` - List of checkpoints:

  * `id` - Unique identifier of the checkpoint.
  * `plan_id` - ID of the backup plan.
  * `created_at` - Creation time of the checkpoint.
  * `status` - Status of the checkpoint.
  * `checkpoint_items` - List of checkpoint items:
    * `id` - Unique identifier of the checkpoint item.
    * `backup_id` - ID of the backup.
    * `chain_id` - ID of the backup chain.
    * `checkpoint_id` - ID of the checkpoint.
    * `created_at` - Creation time of the checkpoint item.
    * `backup_created_at` - Creation time of the backup.
    * `is_incremental` - Whether the backup is incremental.
    * `status` - Status of the checkpoint item.
    * `resource` - List of resource details:
      * `id` - Resource ID.
      * `name` - Resource name.
      * `type` - Resource type.

