---
page_title: "Selectel: selectel_dbaas_roles_v1"
description: |-
  Provides a list of user roles available in Selectel Managed Databases.
---

# selectel\_dbaas\_roles\_v1

Provides a list of user roles available in Managed Databases. For more information about managing users and roles in Managed Databases, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/en/managed-databases/postgresql/manage-users/), [PostgreSQL for 1C](https://docs.selectel.ru/en/managed-databases/postgresql-for-1c/manage-users-1c/), [PostgreSQL TimescaleDB](https://docs.selectel.ru/en/managed-databases/timescaledb/manage-users/), [PostgreSQL PGVector](https://docs.selectel.ru/en/managed-databases/pgvector/manage-users/).

## Example Usage

```terraform
data "selectel_dbaas_roles_v1" "roles" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    datastore_type_id = data.selectel_dbaas_datastore_type_v1.datastore_type_1.datastore_types[0].id
  }
}
```

Filter by `datastore_id` to list only the roles that are available for the type of a specific cluster:

```terraform
data "selectel_dbaas_roles_v1" "roles_2" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  filter {
    datastore_id = selectel_dbaas_postgresql_datastore_v1.datastore_1.id
  }
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/#managed-databases).

* `filter` - (Optional) Values to filter available roles.

  * `datastore_type_id` - (Optional) Unique identifier of the cluster type. You can retrieve information about available cluster types with the [selectel_dbaas_datastore_type_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dbaas_datastore_type_v1) data source.

  * `datastore_id` - (Optional) Unique identifier of the cluster. Used to list the roles available for the cluster type of this cluster. Takes precedence over `datastore_type_id`. Retrieved from the [selectel_dbaas_postgresql_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_postgresql_datastore_v1) resource.

  * `name` - (Optional) Name of the role to search.

## Attributes Reference

* `roles` - List of available roles.

  * `id` - Unique identifier of the role.

  * `datastore_type_id` - Unique identifier of the cluster type for which the role is available.

  * `name` - Name of the role.
