data "selectel_dbaas_roles_v1" "roles" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    datastore_type_id = data.selectel_dbaas_datastore_type_v1.datastore_type_1.datastore_types[0].id
  }
}
