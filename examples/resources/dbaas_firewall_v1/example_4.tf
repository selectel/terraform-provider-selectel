resource "selectel_dbaas_firewall_v1" "firewall_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_redis_datastore_v1.cluster_1.id
  ips          = [ "127.0.0.1" ]
}
