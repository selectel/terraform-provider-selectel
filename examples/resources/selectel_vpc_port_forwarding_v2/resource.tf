resource "selectel_vpc_port_forwarding_v2" "port_forwarding_1" {
  project_id          = selectel_vpc_project_v2.project_1.id
  region              = "ru-6"
  floatingip_id       = selectel_vpc_floatingip_v2.floatingip_1.id
  internal_port_id    = selectel_vpc_port_v2.port_1.id
  internal_ip_address = "192.168.199.50"
  internal_port       = 80
  external_port       = 8080
  protocol            = "tcp"
}
