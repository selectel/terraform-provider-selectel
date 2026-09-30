data "selectel_mks_feature_gates_v1" "fg" {
  project_id = selectel_vpc_project_v2.project_1.id
  region = "ru-3"
}
