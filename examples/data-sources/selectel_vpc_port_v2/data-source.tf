data "selectel_vpc_port_v2" "port_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  fixed_ip   = "192.168.199.50"
}
