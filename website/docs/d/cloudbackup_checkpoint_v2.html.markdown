---
layout: "selectel"
page_title: "Selectel: selectel_cloudbackup_checkpoint_v2"
sidebar_current: "docs-selectel-datasource-cloudbackup-checkpoint-v2"
description: |-
  Provides a list of backup checkpoints for Selectel Backups in the Cloud.
---

# selectel\_cloudbackup\_checkpoint\_v2

Provides a list of created backups for Selectel Backups in the Cloud. For more information about backups, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/backups/about-backups/).

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

* `region` - (Required) Pool where the backup plan is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `filter` - (Optional) Values to filter available checkpoints.

  * `plan_name` - (Optional) Name of the backup plan to search.
  * `volume_name` - (Optional) Name of the volume to search.

## Attributes Reference

* `checkpoints` - List of available checkpoints:

  * `id` - Unique identifier of the checkpoint.
  * `plan_id` - Unique identifier of the backup plan.
  * `created_at` - Time when the checkpoint was created.
  * `status` - Status of the checkpoint.
  * `checkpoint_items` - List of checkpoint items:
    * `id` - Unique identifier of the checkpoint item.
    * `backup_id` - Unique identifier of the backup.
    * `chain_id` - Uniquer identifier of the backup chain.
    * `checkpoint_id` - Uniquer identifier of the checkpoint.
    * `created_at` - Time when the checkpoint item was created.
    * `backup_created_at` - Time when the backup was created.
    * `is_incremental` - Shows whether the backup is incremental.
    * `status` - Status of the checkpoint item.
    * `resource` - List of resource details:
      * `id` - Unique identifier of the resource.
      * `name` - Resource name.
      * `type` - Resource type.

