resource "selectel_vpc_router_route_v2" "route_1" {
  project_id       = selectel_vpc_project_v2.project_1.id
  region           = "ru-6"
  router_id        = selectel_vpc_router_v2.router_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.10"
  depends_on       = [selectel_vpc_router_interface_v2.interface_1]
}
