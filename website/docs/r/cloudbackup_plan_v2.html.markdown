---
layout: "selectel"
page_title: "Selectel: selectel_cloudbackup_plan_v2"
sidebar_current: "docs-selectel-resource-cloudbackup-plan-v2"
description: |-
  Creates and manages a backup plan for Selectel Scheduled Backup service.
---

# selectel\_cloudbackup\_plan\_v2

Creates and manages a backup plan for Selectel Scheduled Backup service. For more information about backup plans, see the [official Selectel documentation](https://docs.selectel.ru/en/api/scheduled-backups/).

## Example Usage

```hcl
resource "selectel_cloudbackup_plan_v2" "plan_1" {
  project_id          = selectel_vpc_project_v2.project_1.id
  region              = "ru-3"
  name                = "my-backup-plan"
  backup_mode         = "full"
  description         = "Nightly backup plan"
  full_backups_amount = 7
  schedule_type       = "crontab"
  schedule_pattern    = "0 0 * * *"
  resources{
    resource {
        id   = "d63dcb8b-77bb-4741-b7dc-1c03c853de12"
        name = "my-volume-1"
        type = "OS::Cinder::Volume"
      }
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the backup plan is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `name` - (Required) Name of the backup plan.

* `backup_mode` - (Optional) Backup mode used for the plan. Available values are `full` and `frequency`. The default value is `full`. Learn more about [backup modes](https://docs.selectel.ru/en/cloud-servers/backups/about-backups/). 

* `description` - (Optional) Description of the backup plan.

* `full_backups_amount` - (Required) Maximum number of backups to save in a full plan or full backups in a frequency plan.

* `schedule_type` - (Optional) Backup scheduling type. Available values are `calendar` and `crontab`. Learn more about [schedule types](https://docs.selectel.ru/en/cloud-servers/backups/create-backup/#configure-scheduled-backups).

* `schedule_pattern` - (Optional) Backup scheduling pattern. The default value is `0 0 * * *`.

* `resources` - (Required) List of resources to back up according to the backup plan. The only available type of resources is a volume. You can add multiple volumes – each volume in a separate block.

  * `resource` - (Required) List of resource objects:
    * `id` - (Required) UUID of the backed up resource.
    * `name` - (Required) Name of the backed up resource.
    * `type` - (Required) Type of the resource to back up. The only available value is `"OS::Cinder::Volume"`.