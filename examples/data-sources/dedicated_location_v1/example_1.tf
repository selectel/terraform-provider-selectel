data "selectel_dedicated_location_v1" "server_location" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name = "SPB-2"
  }
}
