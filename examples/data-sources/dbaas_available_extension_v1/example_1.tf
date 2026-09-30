data "selectel_dbaas_available_extension_v1" "available_extension_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}
