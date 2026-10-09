data "selectel_dbaas_user_setting_parameter_v1" "user_setting_parameter_2" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"

  filter {
    datastore_id = selectel_dbaas_postgresql_datastore_v1.datastore_1.id
  }
}
