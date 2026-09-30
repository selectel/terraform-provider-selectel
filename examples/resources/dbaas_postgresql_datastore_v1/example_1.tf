resource "selectel_dbaas_postgresql_datastore_v1" "cluster_1" {
  name       = "cluster-1"
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  type_id    = data.selectel_dbaas_datastore_type_v1.datastore_type_1.datastore_types[0].id
  subnet_id  = selectel_vpc_subnet_v2.subnet.subnet_id
  node_count = 3
  flavor {
    vcpus     = 4
    ram       = 4096
    disk      = 32
    disk_type = "network-ultra"
  }
  pooler {
    mode = "transaction"
    size = 50
  }
  security_groups = ["796f1f0a-d97d-4a8e-904e-4fd5ef57465c", "b9c2e73d-a6c5-4def-994d-ce85e3ce98d3"]
}
