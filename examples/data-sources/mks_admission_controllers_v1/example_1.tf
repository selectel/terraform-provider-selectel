data "selectel_mks_admission_controllers_v1" "admission_controllers_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region = "ru-3"
}
