# Router interface attached through a subnet
resource "selectel_vpc_router_interface_v2" "interface_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  router_id  = selectel_vpc_router_v2.router_1.id
  subnet_id  = selectel_vpc_network_subnet_v2.subnet_1.id
}
