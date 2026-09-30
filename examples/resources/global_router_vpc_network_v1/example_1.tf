resource "selectel_global_router_vpc_network_v1" "global_router_vpc_network_1" {
  router_id     = selectel_global_router_router_v1.global_router_1.id
  zone_id       = data.selectel_global_router_zone_v1.zone_1.id
  os_network_id = data.openstack_networking_network_v2.network_1.id
  project_id    = selectel_vpc_project_v2.project_1.id
  name          = "my_super_vpc_net"
  tags          = ["blue", "red"]
}
