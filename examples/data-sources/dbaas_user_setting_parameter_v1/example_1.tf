data "selectel_dbaas_user_setting_parameter_v1" "user_setting_parameter_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"

  filter {
    name = "statement_timeout"
  }
}
