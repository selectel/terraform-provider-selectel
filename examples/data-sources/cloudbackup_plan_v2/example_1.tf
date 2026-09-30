data "selectel_cloudbackup_plan_v2" "plan_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    name        = "my-backup-plan"
    volume_name = "my-volume"
    status      = "started"
  }
}
