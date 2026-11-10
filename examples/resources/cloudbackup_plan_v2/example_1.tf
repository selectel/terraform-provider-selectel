resource "selectel_cloudbackup_plan_v2" "plan_1" {
  project_id          = selectel_vpc_project_v2.project_1.id
  region              = "ru-3"
  name                = "my-backup-plan"
  backup_mode         = "full"
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
