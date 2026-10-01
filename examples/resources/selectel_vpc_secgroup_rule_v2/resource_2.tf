# Rule that allows outbound traffic to another security group
resource "selectel_vpc_secgroup_rule_v2" "rule_2" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-6"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_1.id
  direction         = "egress"
  ethertype         = "IPv4"
  remote_group_id   = selectel_vpc_secgroup_v2.secgroup_2.id
}
