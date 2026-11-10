resource "selectel_dbaas_kafka_acl_v1" "acl_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_kafka_datastore_v1.cluster_1.id
  pattern      = "topic"
  pattern_type = "prefixed"
  allow_read   = true
  allow_write  = true
}
