resource "selectel_dbaas_postgresql_extension_v1" "extension_1" {
  project_id                  = selectel_vpc_project_v2.project_1.id
  region                      = "ru-3"
  datastore_id                = selectel_dbaas_postgresql_datastore_v1.cluster_1.id
  database_id                 = selectel_dbaas_postgresql_database_v1.database_1.id
  available_extension_id      = data.selectel_dbaas_available_extension_v1.ae.available_extensions[0].id
}
