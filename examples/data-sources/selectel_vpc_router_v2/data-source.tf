data "selectel_vpc_router_v2" "router_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name       = "router-1"
}
