---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_server_v1"
sidebar_current: "docs-selectel-resource-dedicated-server-v1"
description: |-
  Creates and manages a server in Selectel Dedicated Servers using public API v1.
---

# selectel\_dedicated\_server\_v1

Creates and manages a server in Selectel Dedicated Servers using public API v1. For more information about Dedicated Servers, see the [official Selectel documentation](https://docs.selectel.ru/en/dedicated/).

## Example usage

### Server with one RAID

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
```

### Server with multiple RAID

```hcl
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
```

### Server with single disk configuration without RAID

```hcl
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
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.  Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects).

* `configuration_id` - (Required) Unique identifier of the server configuration. Retrieved from the [dedicated_configuration_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/dedicated_configuration_v1) data source.

* `location_id` - (Required) Pool where the server is located. Retrieved from the [dedicated_location_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/dedicated_location_v1) data source.

* `os_id` - (Required) Unique identifier of the operating system to install. Changing this installs new os on a new server.  Installing a new operating system deletes all data on the server.  Retrieved from the [dedicated_os_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/dedicated_os_v1) data source.

* `price_plan_name` - (Required) The name of the price plan. Available tariff plans are `1 day`, `1 month`, `3 months`, `6 months`, `12 months`, and `12 months • monthly payment`. Learn more about tariff plans in the [Payment model and prices of a dedicated server](https://docs.selectel.ru/en/dedicated/about/payment).

* `os_password` — (Optional) Password for the operating system user.
After the operating system is installed, a password will be generated to connect to the server.
To get the password, in the [Control panel](https://my.selectel.ru/servers), go to **Products** → **Dedicated Servers** → Server page → **Operating System** tab → in the **Password** field.
The password is available for 24 hours from the start of the operating system installation or configuration change.
If you forget the server password, you can [reset and restore it](https://github.com/dedicated/troubleshooting/recover-password/).

* `user_data` - (Optional) Custom configuration settings that automate perform common tasks or execute server setup scripts, reducing the time required to configure and deploy your infrastructure. Learn more about user data in the [User data on a dedicated server](https://docs.selectel.ru/en/dedicated/manage/user-data).

* `ssh_key` - (Optional) The public SSH key to be added to the server.

* `ssh_key_name` - (Optional) Name of an existing SSH key to be added to the server. Learn more about add a public SSH key to the SSH key repository in the [Create and host an SSH key on a dedicated server](https://docs.selectel.ru/en/dedicated/manage/create-and-place-ssh-key).

* `partitions_config` - (Optional) Configuration for disk partitions. Learn more about disk partitioning in the [Install the OS by auto-installation](https://docs.selectel.ru/en/dedicated/manage/autoinstall-os/#partition-disks).
  * `soft_raid_config` - (Optional) Configuration for software RAID. You can add configurations for multiple RAID arrays – each configuration in a separate block. Creating multiple arrays requires 4 or more disks.
    * `name` - (Required) Name of the RAID array.
    * `level` - (Required) RAID level. Available values are raid0, raid1, and raid10.
    * `disk_type` - (Required) Type of disks to use in the RAID. Available values are SSD, NVMe, HDD.
    * `count` - (Optional) Number of disks to use in the RAID array. If not specified, the minimum number of disks depends on the selected RAID level: 2 — for RAID0 and RAID1, 4 — for RAID10.
  * `disk_partitions` - (Optional) Configuration for disk partitions. You can add configurations for multiple disk partitions – each configuration in a separate block.
    * `mount` - (Required) Mount point for a partition. Required mount points are /, /boot, swap. You can use additional mount points, for example, /data.
    * `size` - (Optional) Size of the partition in GB. Use either size or size_percent. To use all the remaining space for the partition, specify -1.
    * `size_percent` - (Optional) Size of the partition in percent (0-100). Use either size or size_percent.
    * `raid` - (Optional) RAID array name on which to create the partition. Required only when using RAID.
    * `disk_name` - (Optional) Name of the disk on which to create the partition. RAID is not used.
    * `fs_type` - (Optional) Filesystem type for the partition. Available file system types are swap, ext4, ext3, and xfs. The default value is ext4 for /, ext3 for /boot, swap for swap partition.
  * `disk_config` - (Optional) Configuration for individual disks without RAID. You can add configurations for multiple disks – each configuration in a separate block.
    * `name` - (Required) Name of the disk to reference in `disk_partitions`.
    * `disk_type` - (Required) Type of disks to use in the RAID. Available values are SSD, NVMe, HDD.

* `public_subnet_id` - (Optional) Unique identifier of the public subnet to connect to the server. Retrieved from the [selectel_dedicated_public_subnet_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dedicated_public_subnet_v1) data source.

* `public_subnet_ip` - (Optional) Public IP address to assign to the server within the public subnet.

* `private_subnet_id` - (Optional) Unique identifier of the private subnet to connect to the server. Changing this deletes the existing server and creates a new server. Retrieved from the [selectel_dedicated_private_subnet_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/dedicated_private_subnet_v1) data source.

* `private_subnet_ip` - (Optional) Private IP address to assign to the server within the private subnet. Must be used together with private_subnet_id.

* `add_private_vlan` - (Optional) Adds a private VLAN to the server. Boolean flag, the default value is false.

* `os_host_name` - (Optional) Hostname for the server.

* `power_state` - (Optional) Power state of the server. Available values are on, off, and reboot. The default value is on. Cannot be set during server creation — servers are always created in the on state. Use power_state only for updating an existing server's power state. Changing power_state is mutually exclusive with changes of arguments: os_id, os_password, ssh_key, ssh_key_name, partitions_config, user_data, os_host_name and force_update_additional_params.

* `force_update_additional_params` - (Optional) Enables or disables update of operating system parameters without changing os_id. The operating system parameters are os_password, user_data, ssh_key, ssh_key_name, partitions_config, and os_host_name. After updating operating system parameters, a new operating system will be installed. Installation of a new operating system will delete all data on the server.

* `timeouts` — (Optional) Timeout values.
The default values are the following:
  * create = "80m",
  * update = "20m",
  * delete = "5m".

## Attributes Reference

* `id` - Unique identifier of the server.
* `public_ip` - Public IP address of the server.
* `private_ip` - Private IP address of the server.
* `private_vlan` - Unique identifier of the VLAN of the private subnet assigned to the server. Returned when a private VLAN is configured or created for the server.

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

When importing a server, Terraform auto-generates the names of the following objects:

* RAID array names — auto-generated based on the RAID level and disk type. For example, new-raid1, new-raid0.

* disk names — auto-generated based on the disk type. For example, disk-ssd-1, disk-hdd-1.

You may need to update your configuration to match these names or use terraform import followed by terraform state show to see the imported configuration.
