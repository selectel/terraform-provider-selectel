data "selectel_dedicated_configuration_v1" "server_config" {
  project_id = selectel_vpc_project_v2.project_1.id

  filter {
    name        = "CL25-NVMe"
    location_id = data.selectel_dedicated_location_v1.server_location.locations[0].id
  }
}
