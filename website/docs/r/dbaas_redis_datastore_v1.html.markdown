---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_redis_datastore_v1"
sidebar_current: "docs-selectel-resource-dbaas-redis-datastore-v1"
description: |-
  Manages a V1 Redis datastore resource within Selectel Managed Databases Service.
---

# selectel\_dbaas\_redis\_datastore\_v1

Manages a V1 Redis datastore resource within Selectel Managed Databases Service.

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
    engine  = "redis"
    version = "6"
  }
}

data "selectel_dbaas_flavor_v1" "flavor" {
    project_id   = "${selectel_vpc_project_v2.project_1.id}"
    region = "ru-3"
    filter {
        datastore_type_id = data.selectel_dbaas_datastore_type_v1.dt_redis.datastore_types[0].id
    }
}

resource "selectel_dbaas_redis_datastore_v1" "datastore_1" {
  name         = "datastore-1"
  project_id   = "${selectel_vpc_project_v2.project_1.id}"
  region       = "ru-3"
  type_id      = data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id
  subnet_id    = "${selectel_vpc_subnet_v2.subnet.subnet_id}"
  node_count   = 3
  flavor_id = data.selectel_dbaas_flavor_v1.flavor.flavors[0].id
  config = {
    maxmemory-policy = "allkeys-lru"
  }
  redis_password = "secret"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name of the datastore.
  Changing this creates a new datastore.

* `project_id` - (Required) An associated Selectel VPC project.
  Changing this creates a new datastore.

* `region` - (Required) A Selectel VPC region of where the datastore is located.
  Changing this creates a new datastore.

* `subnet_id` - (Required) Associated OpenStack Networking service subnet ID.
  Changing this creates a new datastore.

* `type_id` - (Required) The datastore type for the datastore.
  Changing this creates a new datastore.

* `node_count` - (Required) Number of nodes to create for the datastore.

* `flavor_id` - (Required) Flavor identifier for the datastore.

* `firewall` - (Optional) List of the ips to allow access from.

* `restore` - (Optional) Restore parameters for the datastore. It's a complex value. See description below.
  Changing this creates a new datastore.

* `config` - (Optional) Configuration parameters for the datastore.

* `redis_password` - (Required) Password for the Redis datastore

**restore**

- `datastore_id` - (Optional) - Datastore ID to restore from.
- `target_time` - (Optional) - Restore by the target time.

## Attributes Reference

The following attributes are exported:

* `status` - Shows the current status of the datastore.

* `connections` - Shows DNS connection strings for the datastore.

## Import

Datastore can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN SEL_PROJECT_ID=SELECTEL_VPC_PROJECT_ID SEL_REGION=SELECTEL_VPC_REGION terraform import selectel_dbaas_datastore_v1.datastore_1 b311ce58-2658-46b5-b733-7a0f418703f2
```
