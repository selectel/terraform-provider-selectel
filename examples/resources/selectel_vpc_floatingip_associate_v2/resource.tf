resource "selectel_vpc_floatingip_associate_v2" "association_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-6"
  floating_ip = selectel_vpc_floatingip_v2.floatingip_1.floating_ip_address
  port_id     = selectel_vpc_port_v2.port_1.id
  depends_on  = [selectel_vpc_router_interface_v2.interface_1]
}
