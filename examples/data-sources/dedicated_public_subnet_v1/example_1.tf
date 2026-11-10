data "selectel_dedicated_public_subnet_v1" "public_subnets" {
  project_id = selectel_vpc_project_v2.project_1.id
  filter {
    location_id = data.selectel_dedicated_location_v1.server_location.locations[0].id
    subnet = "192.168.1.0/29"
    ip = "192.168.1.3"
  }
}
