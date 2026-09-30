resource "selectel_dedicated_server_v1" "server_1" {
  project_id = selectel_vpc_project_v2.project_1.id

  configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
  location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  os_id            = data.selectel_dedicated_os_v1.server_os.os[0].id
  price_plan_name  = "1 day"

  os_host_name     = "Turing"
  public_subnet_id = data.selectel_dedicated_public_subnet_v1.subnets.subnets[0].id
  # public_subnet_ip = data.selectel_dedicated_public_subnet_v1.subnets.subnets[0].ip
  private_subnet_id = var.private_subnet_id  # Optional: Private subnet ID
  private_subnet_ip = "192.168.100.10"       # Optional: Specific private IP
  ssh_key_name     = "deploy-ed25519"
  os_password      = "Passw0rd!"
  user_data        = file("init-script-dir/init.sh")

  partitions_config {
    soft_raid_config {
      name      = "first-raid"
      level     = "raid1"
      disk_type = "SSD NVMe M.2"
    }

    disk_partitions {
      mount = "/boot"
      size  = 1
      raid  = "first-raid"
    }
    disk_partitions {
      mount        = "swap"
      size_percent = 10.5
      raid         = "first-raid"
    }
    disk_partitions {
      mount = "/"
      size  = -1
      raid  = "first-raid"
    }
    disk_partitions {
      mount   = "second_folder"
      size    = 400
      raid    = "first-raid"
      fs_type = "xfs"
    }
  }

  timeouts {
    create = "80m"
    update = "20m"
    delete = "5m"
  }
}
