resource "selectel_global_router_vpc_subnet_v1" "global_router_vpc_subnet_1" {
  network_id        = selectel_global_router_vpc_network_v1.global_router_vpc_network_1.id
  os_subnet_id      = data.openstack_networking_subnet_v2.subnet_1.id
  cidr              = "10.10.10.0/24"
  gateway           = "10.10.10.13"
  service_addresses = ["10.10.10.253", "10.10.10.254"]
  name              = "my_super_vpc_subnet"
  tags              = ["blue", "red"]
}
