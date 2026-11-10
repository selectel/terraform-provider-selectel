resource "selectel_vpc_project_v2" "project_1" {
  name = "project1"
  quotas {
    resource_name = "compute_cores"
    resource_quotas {
      region = "ru-3"
      zone   = "ru-3a"
      value  = 12
    }
  }
  quotas {
    resource_name = "compute_ram"
    resource_quotas {
      region = "ru-3"
      zone   = "ru-3a"
      value  = 20480
    }
  }
  quotas {
    resource_name = "volume_gigabytes_fast"
    resource_quotas {
      region = "ru-3"
      zone   = "ru-3a"
      value  = 100
    }
  }
}
