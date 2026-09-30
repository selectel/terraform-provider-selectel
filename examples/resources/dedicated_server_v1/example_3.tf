resource "selectel_dedicated_server_v1" "server_single_disk" {
  project_id = selectel_vpc_project_v2.project_1.id

  configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
  location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  os_id            = data.selectel_dedicated_os_v1.server_os.os[0].id
  price_plan_name  = "1 day"

  partitions_config {
    disk_config {
      name      = "system-disk"
      disk_type = "SSD NVMe"
    }
    disk_config {
      name      = "data-disk"
      disk_type = "HDD SATA"
    }

    disk_partitions {
      mount     = "/boot"
      size      = 1
      disk_name = "system-disk"
    }
    disk_partitions {
      mount     = "/"
      size      = -1
      disk_name = "system-disk"
    }
    disk_partitions {
      mount     = "/data"
      size      = -1
      disk_name = "data-disk"
      fs_type   = "xfs"
    }
  }
}
