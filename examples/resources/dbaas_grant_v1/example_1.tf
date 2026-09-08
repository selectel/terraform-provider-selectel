resource "selectel_dbaas_grant_v1" "grant_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_postgresql_datastore_v1.cluster_1.id
  database_id  = selectel_dbaas_postgresql_database_v1.database_1.id
  user_id      = selectel_dbaas_user_v1.user_1.id
}
