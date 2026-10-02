data "selectel_dbaas_flavor_v2" "flavor" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    datastore_type_id = data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id
    allowed_role      = "KEEPER"
    fl_size           = "STANDARD"
  }
}
