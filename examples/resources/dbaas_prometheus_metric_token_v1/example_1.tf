resource "selectel_dbaas_prometheus_metric_token_v1" "token_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "token"
}
