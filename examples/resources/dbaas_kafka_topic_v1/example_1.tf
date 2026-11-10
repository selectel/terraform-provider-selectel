resource "selectel_dbaas_kafka_topic_v1" "topic_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_kafka_datastore_v1.cluster_1.id
  name         = "topic"
  partitions   = 1
}
