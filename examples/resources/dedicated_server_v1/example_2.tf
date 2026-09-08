resource "selectel_dedicated_server_v1" "server_multi_raid" {
  project_id = selectel_vpc_project_v2.project_1.id

  configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
  location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  os_id            = data.selectel_dedicated_os_v1.server_os.os[0].id
  price_plan_name  = "1 day"

  partitions_config {
    soft_raid_config {
      name      = "boot-raid"
      level     = "raid1"
      disk_type = "SSD NVMe"
      count     = 2
    }

    soft_raid_config {
      name      = "data-raid"
      level     = "raid0"
      disk_type = "SSD NVMe"
      count     = 2
    }

    disk_partitions {
      mount = "/boot"
      size  = 1
      raid  = "boot-raid"
    }
    disk_partitions {
      mount = "/"
      size  = -1
      raid  = "boot-raid"
    }
    disk_partitions {
      mount = "/data"
      size  = -1
      raid  = "data-raid"
      fs_type = "xfs"
    }
  }
}
