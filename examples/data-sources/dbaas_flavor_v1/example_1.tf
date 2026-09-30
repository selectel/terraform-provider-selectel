data "selectel_dbaas_flavor_v1" "flavor" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}
