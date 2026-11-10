data "selectel_dbaas_datastore_type_v1" "datastore_type_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine  = "mysql_native"
    version = "8"
  }
}
