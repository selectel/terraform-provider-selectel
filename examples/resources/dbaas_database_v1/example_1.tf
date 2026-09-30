resource "selectel_vpc_project_v2" "project_1" {
}

resource "selectel_vpc_subnet_v2" "subnet" {
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
  filter {
    engine  = "postgresql"
    version = "12"
  }
}

resource "selectel_dbaas_datastore_v1" "datastore_1" {
  name         = "datastore-1"
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
  type_id      = data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id
  subnet_id    = "${selectel_vpc_subnet_v2.subnet.subnet_id}"
  node_count   = 3
  flavor {
    vcpus = 4
    ram   = 4096
    disk  = 32
  }
  pooler {
    mode = "transaction"
    size = 50
  }
}

resource "selectel_dbaas_user_v1" "user_1" {
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
  datastore_id = "${selectel_dbaas_datastore_v1.datastore_1.id}"
  name         = "user"
  password     = "secret"
}

resource "selectel_dbaas_database_v1" "database_1" {
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
  datastore_id = "${selectel_dbaas_datastore_v1.datastore_1.id}"
  owner_id     = "${selectel_dbaas_user_v1.user_1.id}"
  name         = "db"
  lc_ctype     = "ru_RU.utf8"
  lc_collate   = "ru_RU.utf8"
}
