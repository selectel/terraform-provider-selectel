resource "selectel_dbaas_mysql_database_v1" "database_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_mysql_datastore_v1.cluster_1.id
  name         = "database_1"
}
