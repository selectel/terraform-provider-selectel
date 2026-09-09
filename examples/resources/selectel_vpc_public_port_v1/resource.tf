resource "selectel_vpc_public_port_v1" "port_1" {
  project_id         = selectel_vpc_project_v2.project_1.id
  region             = "ru-6"
  description        = "my-own-direct-public-ip-port"
  admin_state_up     = false
  security_group_ids = [openstack_networking_secgroup_v2.sg_1.id]
}
