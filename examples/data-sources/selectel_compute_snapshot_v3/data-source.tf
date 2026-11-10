data "selectel_compute_snapshot_v3" "snapshot_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name       = "snapshot-1"
  volume_id  = selectel_compute_volume_v3.volume_1.id
}
