data "selectel_cloudbackup_checkpoint_v2" "checkpoint_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    plan_name   = "my-backup-plan"
    volume_name = "my-volume"
  }
}
