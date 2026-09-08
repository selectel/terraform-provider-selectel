resource "selectel_vpc_floatingip_v2" "floatingip_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-1"
}
