resource "selectel_global_router_dedicated_network_v1" "global_router_dedicated_network_1" {
  router_id = selectel_global_router_router_v1.global_router_1.id
  zone_id   = data.selectel_global_router_zone_v1.zone_1.id
  vlan      = "1234"
  name      = "my_super_dedicated_net"
  tags      = ["blue", "red"]
}
