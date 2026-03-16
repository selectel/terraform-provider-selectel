---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_server_v1"
sidebar_current: "docs-selectel-resource-dedicated-server-v1"
description: |-
  Creates and manages a server in Selectel Dedicated Servers.
---

# selectel\_dedicated\_server\_v1

Creates and manages a server in Selectel Dedicated Servers.

## Example usage

```hcl
resource "selectel_dedicated_server_v1" "server_1" {
  project_id = selectel_vpc_project_v2.project_1.id

  configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
  location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  os_id            = data.selectel_dedicated_os_v1.server_os.os[0].id
  price_plan_name  = "1 day"

  os_host_name     = "Turing"
  public_subnet_id = data.selectel_dedicated_public_subnet_v1.subnets.subnets[0].id
  # public_subnet_ip = data.selectel_dedicated_public_subnet_v1.subnets.subnets[0].ip
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

  # Optional: You can choose your own timeout values or remove them.
  #
  # Current values represent default values.
  timeouts {
    create = "80m"
    update = "20m"
    delete = "5m"
  }
}

# Multiple RAID arrays example (RAID1 + RAID0)
resource "selectel_dedicated_server_v1" "server_multi_raid" {
  project_id = selectel_vpc_project_v2.project_1.id

  configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
  location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  os_id            = data.selectel_dedicated_os_v1.server_os.os[0].id
  price_plan_name  = "1 day"

  partitions_config {
    # RAID1 for boot and root (2 disks)
    soft_raid_config {
      name      = "boot-raid"
      level     = "raid1"
      disk_type = "SSD NVMe"
      count     = 2
    }

    # RAID0 for data (2 disks)
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

# Single disk configuration without RAID
resource "selectel_dedicated_server_v1" "server_single_disk" {
  project_id = selectel_vpc_project_v2.project_1.id

  configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
  location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  os_id            = data.selectel_dedicated_os_v1.server_os.os[0].id
  price_plan_name  = "1 day"

  partitions_config {
    # Define individual disks
    disk_config {
      name      = "system-disk"
      disk_type = "SSD NVMe"
    }
    disk_config {
      name      = "data-disk"
      disk_type = "HDD SATA"
    }

    # Partitions on specific disks
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
```

# Power management example
resource "selectel_dedicated_server_v1" "server_power" {
  project_id = selectel_vpc_project_v2.project_1.id

  configuration_id = data.selectel_dedicated_configuration_v1.server_config.configurations[0].id
  location_id      = data.selectel_dedicated_location_v1.server_location.locations[0].id
  os_id            = data.selectel_dedicated_os_v1.server_os.os[0].id
  price_plan_name  = "1 day"

  # Power management only
  power_state = "on"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.  Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `configuration_id` - (Required) Unique identifier of the server configuration. Retrieved from the [dedicated_configuration_v1]((https://registry.terraform.io/providers/selectel/selectel/latest/docs/dedicated_configuration_v1) data source.

* `location_id` - (Required) Pool where the server is located. Retrieved from the [dedicated_location_v1]((https://registry.terraform.io/providers/selectel/selectel/latest/docs/dedicated_location_v1) data source.

* `os_id` - (Required) Unique identifier of the operating system to install. Changing this installs new os on a new server.  Installing new os will delete all data on the server.  Retrieved from the [dedicated_os_v1]((https://registry.terraform.io/providers/selectel/selectel/latest/docs/dedicated_os_v1) data source.

* `price_plan_name` - (Required) The name of the price plan. Available tariff plans are `1 day`, `1 month`, `3 months`, `6 months`, `12 months`, and `12 months • monthly payment`. Learn more about tariff plans in the [Payment model and prices of a dedicated server]((https://docs.selectel.ru/en/dedicated/about/payment/).

* `os_password` - (Optional) Password for the OS user.

* `user_data` - (Optional) These are custom configuration settings that automatically perform common tasks or run server setup scripts, reducing the time it takes to configure and deploy your infrastructure. Learn more about user data in the [User data on a dedicated server]((https://docs.selectel.ru/en/dedicated/manage/user-data/).

* `ssh_key` - (Optional) The public SSH key to be added to the server.

* `ssh_key_name` - (Optional) The name of an existing SSH key to be added to the server. Learn more about add a public SSH key to the SSH key repository in the [Create and host an SSH key on a dedicated server]((https://docs.selectel.ru/en/dedicated/manage/create-and-place-ssh-key/).

* `partitions_config` - (Optional) Configuration for disk partitions. Learn more about disk partitioning in the [Install the OS by auto-installation]((https://docs.selectel.ru/en/dedicated/manage/autoinstall-os/#partition-disks).
  * `soft_raid_config` - (Optional) Configuration for software RAID. Can be specified multiple times to create multiple RAID arrays (requires 4+ disks for multiple arrays).
    * `name` - (Required) Name of the RAID array.
    * `level` - (Required) RAID level. Valid values are `raid0`, `raid1`, and `raid10`.
    * `disk_type` - (Required) Type of disks to use in the RAID (e.g., `SSD NVMe`, `HDD SATA`).
    * `count` - (Optional) Number of disks to use in the RAID array. If not specified, defaults to the minimum required for the RAID level (2 for raid0/raid1, 4 for raid10).
  * `disk_partitions` - (Optional) List of disk partitions. Can be specified multiple times.
    * `mount` - (Required) Mount point for the partition (e.g., `/`, `/boot`, `swap`, `/data`).
    * `size` - (Optional) Size of the partition in GB. Use only `size` or `size_percent`, not both. Use `-1` for all remaining space.
    * `size_percent` - (Optional) Size of the partition in percent (0-100). Use only `size` or `size_percent`, not both.
    * `raid` - (Optional) The RAID array name to create the partition on. Required when using RAID.
    * `disk_name` - (Optional) The name of the disk (from `disk_config`) to create the partition on. Required when not using RAID.
    * `fs_type` - (Optional) Filesystem type for the partition. Available file system types are `swap`, `ext4`, `ext3`, and `xfs`. Defaults to `ext4` (or `ext3` for `/boot`, `swap` for swap partition).
  * `disk_config` - (Optional) Configuration for individual disks (used when not using RAID). Can be specified multiple times.
    * `name` - (Required) Name of the disk to reference in `disk_partitions`.
    * `disk_type` - (Required) Type of the disk (e.g., `SSD NVMe`, `HDD SATA`).

* `public_subnet_id` - (Optional) ID of the public subnet to connect the server to. If id is set, the first free subnet address wil be used.

* `public_subnet_ip` - (Optional) Public IP to use. Can be set instead of `public_subnet_id`.

* `os_host_name` - (Optional) Hostname for the server.

* `power_state` - (Optional) Power state of the server. Valid values are `on`, `off`, and `reboot`. **Note:** This field cannot be set during server creation - servers are always created in the "on" state. Use `power_state` only for updating an existing server's power state. Changing `power_state` is mutually exclusive with other configuration changes (`os_id`, `os_password`, `ssh_key`, `ssh_key_name`, `partitions_config`, `user_data`, `os_host_name`, `force_update_additional_params`). This validation occurs at plan time to prevent state corruption. Setting `power_state = "off"` prevents OS installation operations - the server must be powered on first.

* `force_update_additional_params` - (Optional) Enable or disable update for additional os params (os_password, user_data, ssh_key, ssh_key_name, partitions_config, os_host_name) without changing os_id. NOTE: installing new os will delete all data on the server.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier of the server.

## Import

You can import a server:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<selectel_project_id>
terraform import selectel_dedicated_server_v1.server_1 <server_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/servers), go to **Servers and colocation** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<server_id>` — Unique identifier of the server.

### Import notes

* After importing a server with a custom disk partition configuration (`partitions_config`), Terraform will read the actual partition layout from the API and populate the state accordingly.

* When importing a server with software RAID configurations, the RAID array names will be auto-generated (e.g., `new-raid1`, `new-raid0`) based on the RAID level and disk type. You may need to update your configuration to match these names or use `terraform import` followed by `terraform state show` to see the imported configuration.

* Disk names in `disk_config` are also auto-generated during import (e.g., `disk-ssd-1`, `disk-hdd-1`) based on the disk type.
