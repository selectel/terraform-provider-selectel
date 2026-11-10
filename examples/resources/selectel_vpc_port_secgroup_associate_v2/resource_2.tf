# Security groups that replace the existing groups of the port
resource "selectel_vpc_port_secgroup_associate_v2" "associate_2" {
  project_id         = selectel_vpc_project_v2.project_1.id
  region             = "ru-6"
  port_id            = data.selectel_vpc_port_v2.port_1.id
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_1.id]
  enforce            = true
}
