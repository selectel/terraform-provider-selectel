resource "selectel_global_router_dedicated_subnet_v1" "global_router_dedicated_subnet_1" {
  network_id        = selectel_global_router_dedicated_network_v1.global_router_dedicated_network_1.id
  cidr              = "10.10.10.0/24"
  gateway           = "10.10.10.13"
  service_addresses = ["10.10.10.253", "10.10.10.254"]
  name              = "my_super_dedicated_subnet"
  tags              = ["blue", "red"]
}
