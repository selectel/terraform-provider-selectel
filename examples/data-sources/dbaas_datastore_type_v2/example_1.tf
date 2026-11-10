data "selectel_dbaas_datastore_type_v2" "dt" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    engine = "clickhouse"
  }
}
