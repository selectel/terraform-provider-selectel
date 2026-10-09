resource "selectel_dbaas_user_v1" "user_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_postgresql_datastore_v1.cluster_1.id
  name         = "user"
  password     = "secret"

  settings = {
    conn_limit         = "20"
    statement_timeout  = "5000"
    login              = "true"
    synchronous_commit = "on"
  }
}

