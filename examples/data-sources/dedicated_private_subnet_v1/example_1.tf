data "selectel_dedicated_private_subnet_v1" "subnet_ds" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    location_id = data.selectel_dedicated_location_v1.server_location.locations[0].id
    vlan        = "1000"
    subnet      = "192.168.100.0/24"
  }
}
