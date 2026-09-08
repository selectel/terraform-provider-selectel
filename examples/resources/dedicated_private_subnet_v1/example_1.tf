resource "selectel_dedicated_private_subnet_v1" "subnet_1" {
  location_id = data.selectel_dedicated_location_v1.server_location.locations[0].id
  vlan        = "1000"
  subnet      = "192.168.100.0/24"
}
