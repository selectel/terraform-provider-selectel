data "selectel_compute_volume_type_v3" "volume_type_1" {
  project_id     = selectel_vpc_project_v2.project_1.id
  region         = "ru-6"
  volume_type_id = "default"
}
