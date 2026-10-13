data "selectel_vpc_network_v2" "network_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name       = "network-1"
}
