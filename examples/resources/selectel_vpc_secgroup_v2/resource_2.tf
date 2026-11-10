# Security group without the default egress rules
resource "selectel_vpc_secgroup_v2" "secgroup_2" {
  project_id           = selectel_vpc_project_v2.project_1.id
  region               = "ru-6"
  name                 = "secgroup-2"
  delete_default_rules = true
}
