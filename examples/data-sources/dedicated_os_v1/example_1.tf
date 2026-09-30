data "selectel_dedicated_os_v1" "server_os" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    name             = "Ubuntu"
    version_value          = "22.04"
    # version_name     = "22.04 LTS"
    configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
    location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  }
}
