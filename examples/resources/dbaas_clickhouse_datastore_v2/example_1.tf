resource "selectel_dbaas_clickhouse_datastore_v2" "cluster_1" {
  name       = "cluster-1"
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  type_id    = data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id
  subnet_id  = selectel_vpc_subnet_v2.subnet.subnet_id
  password   = "secretsecretsecretsecret"

  node_group {
    name       = "keepers"
    role       = "KEEPER"
    node_count = 3
    flavor {
      id   = data.selectel_dbaas_flavor_v2.keeper_flavor.flavors[0].id
      type = "FIXED"
    }
  }

  node_group {
    name           = "shard1"
    role           = "DATA"
    node_count     = 1
    weight         = 100
    has_public_ips = true
    flavor {
      vcpus     = 2
      ram       = 8192
      disk      = 32
      disk_type = "NETWORK_ULTRA"
      type      = "FLEXIBLE"
    }
  }

  node_group {
    name           = "shard2"
    role           = "DATA"
    node_count     = 1
    weight         = 100
    has_public_ips = true
    flavor {
      vcpus     = 2
      ram       = 8192
      disk      = 32
      disk_type = "NETWORK_ULTRA"
      type      = "FLEXIBLE"
    }
  }

  security_groups = ["796f1f0a-d97d-4a8e-904e-4fd5ef57465c", "b9c2e73d-a6c5-4def-994d-ce85e3ce98d3"]
}
