resource "selectel_vpc_network_subnet_route_v2" "route_1" {
  project_id       = selectel_vpc_project_v2.project_1.id
  region           = "ru-6"
  subnet_id        = selectel_vpc_network_subnet_v2.subnet_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.10"
}
