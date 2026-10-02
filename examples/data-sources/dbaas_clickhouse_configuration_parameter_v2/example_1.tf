data "selectel_dbaas_clickhouse_configuration_parameter_v2" "configuration_parameter_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}
