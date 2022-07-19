---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_postgresql_database_v1"
sidebar_current: "docs-selectel-resource-dbaas-postgresql-database-v1"
description: |-
  Manages a V1 PostgreSQL database resource within Selectel Managed Databases Service.
---

# selectel\_dbaas\_postgresql\_database\_v1

Manages a V1 PostgreSQL database resource within Selectel Managed Databases Service.

## Example usage

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  auto_quotas = true
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

resource "selectel_dbaas_postgresql_datastore_v1" "datastore_1" {
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
  datastore_id = "${selectel_dbaas_postgresql_datastore_v1.datastore_1.id}"
  name         = "user"
  password     = "secret"
}

resource "selectel_dbaas_postgresql_database_v1" "database_1" {
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
  datastore_id = "${selectel_dbaas_postgresql_datastore_v1.datastore_1.id}"
  owner_id     = "${selectel_dbaas_user_v1.user_1.id}"
  name         = "db"
  lc_ctype     = "ru_RU.utf8"
  lc_collate   = "ru_RU.utf8"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name of the database.
  Changing this creates a new database.

* `project_id` - (Required) An associated Selectel VPC project.
  Changing this creates a new database.

* `region` - (Required) A Selectel VPC region of where the database is located.
  Changing this creates a new database.

* `datastore_id` - (Required) An associated datastore.
  Changing this creates a new database.

* `owner_id` - (Optional) Owner of the database. Required only for the PostgreSQL datastore. Can be omitted for the MySQL datastore.

* `lc_collate` - (Optional) A lc_collate option for the PostreSQL datastore.

* `lc_ctype` - (Optional) A lc_ctype option for the PostreSQL datastore.

## Attributes Reference

The following attributes are exported:

* `status` - Shows the current status of the database.

## Import

Database can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN SEL_PROJECT_ID=SELECTEL_VPC_PROJECT_ID SEL_REGION=SELECTEL_VPC_REGION terraform import selectel_dbaas_database_v1.database_1 b311ce58-2658-46b5-b733-7a0f418703f2
```
