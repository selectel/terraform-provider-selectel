---
layout: "selectel"
page_title: "Selectel: selectel_cloudbackup_plan_v2"
sidebar_current: "docs-selectel-datasource-cloudbackup-plan-v2"
description: |-
  Provides a list of backup plans for Selectel Backups in the Cloud.
---

# selectel\_cloudbackup\_plan\_v2

Provides a list of backup plans for Selectel Backups in the Cloud. For more information about backup plans, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/backups/about-backups/).

## Example Usage

```hcl
data "selectel_cloudbackup_plan_v2" "plan_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    name        = "my-backup-plan"
    volume_name = "my-volume"
    status      = "started"
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the backup plan is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `filter` - (Optional) Values to filter backup plans.

  * `name` - (Optional) Name of the backup plan.
  * `volume_name` - (Optional) Name of the volume.
  * `status` - (Optional) Status of the backup plan.

## Attributes Reference

* `plans` - List of backup plans:

  * `list`- Plans list:

    * `id` - Unique identifier of the backup plan.
    * `name` - Name of the backup plan.
    * `description` - Description of the backup plan.
    * `status` - Status of the backup plan.
    * `backup_mode` - Backup mode.
    * `created_at` - Time when the backup plan was created.
    * `full_backups_amount` - Number of full backups.
    * `resources` - List of resources that are backed up according to the backup plan:
      * `id` - Unique identifier of the resource that is backed up according to the backup plan.
      * `name` - Resource name.
      * `type` - Resource type.
    * `schedule_pattern` - Schedule pattern for the backup plan.
    * `schedule_type` - Schedule type for the backup plan.
  
  * `total` - Total number of backup plans.


