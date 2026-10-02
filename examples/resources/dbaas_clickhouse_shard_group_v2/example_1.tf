resource "selectel_dbaas_clickhouse_shard_group_v2" "shard_group_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_clickhouse_datastore_v2.cluster_1.id
  name         = "group-1"
  description  = "First shard group"
  shard_names  = ["shard1", "shard2"]
}
