resource "selectel_vpc_public_port_v1" "port_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
}
