data "selectel_dbaas_configuration_parameter_v1" "configuration_parameter_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}
