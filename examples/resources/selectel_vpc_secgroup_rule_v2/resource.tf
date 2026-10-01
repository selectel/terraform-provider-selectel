# Allowing inbound SSH from one subnet
resource "selectel_vpc_secgroup_rule_v2" "rule_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-6"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "192.168.199.0/24"
}

# Allowing outbound traffic to the members of another group
resource "selectel_vpc_secgroup_rule_v2" "rule_2" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-6"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_1.id
  direction         = "egress"
  ethertype         = "IPv4"
  remote_group_id   = selectel_vpc_secgroup_v2.secgroup_2.id
}
