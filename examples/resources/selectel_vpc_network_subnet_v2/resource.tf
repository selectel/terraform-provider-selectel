resource "selectel_vpc_network_subnet_v2" "subnet_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  network_id = selectel_vpc_network_v2.network_1.id
  cidr       = "192.168.199.0/24"
}
