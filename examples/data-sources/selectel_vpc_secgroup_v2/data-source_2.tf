# Default security group of the project
data "selectel_vpc_secgroup_v2" "secgroup_2" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name       = "default"
}
