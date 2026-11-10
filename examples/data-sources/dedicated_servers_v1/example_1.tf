data "selectel_dedicated_servers_v1" "servers" {
  project_id = selectel_vpc_project_v2.project.id
  filter {
    name = "production-web-01"
    ip = "192.168.1.100"
    location_id   = data.selectel_location_v1.server_location.locations[0].id
    configuration = "EL5"
    private_subnet = data.selectel_dedicated_private_subnet_v1.subnets.subnets[0].id
    public_subnet = data.selectel_dedicated_public_subnet_v1.subnets.subnets[0].id
  }
}
