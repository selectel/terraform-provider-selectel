---
layout: "selectel"
page_title: "Selectel: selectel_dedicated_servers_server_v1"
sidebar_current: "docs-selectel-resource-dedicated-servers-server-v1"
description: |-
  Creates and manages a server in Selectel Dedicated Servers.
---

# selectel\_dedicated\_servers\_server\_v1

Creates and manages a server in Selectel Dedicated Servers.

## Example usage

```hcl
resource "selectel_dedicated_dedicated_servers_server_v1" "server_1" {
  project_id = selectel_vpc_project_v2.project_1.id

  configuration_id = data.selectel_dedicated_servers_configuration_v1.server_config.configurations[0].id
  location_id      = data.selectel_dedicated_servers_location_v1.server_location.locations[0].id
  os_id            = data.selectel_dedicated_servers_os_v1.server_os.os[0].id
  price_plan_name  = "1 день"

  os_host_name     = "Turing"
  public_subnet_id = data.selectel_dedicated_servers_public_subnet_v1.subnets.subnets[0].id
  private_subnet   = "192.168.0.0/16"
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
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project.  Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `configuration_id` - (Required) Unique identifier of the server configuration. 

* `location_id` - (Required) Pool where the server is located. 

* `os_id` - (Required) Unique identifier of the operating system to install. Changing this installs new os on a new server. NOTE: installing new os will delete all data on the server.

* `price_plan_name` - (Required) The name of the price plan.

* `os_password` - (Optional) Password for the OS user.

* `user_data` - (Optional) These are custom configuration settings that automatically perform common tasks or run server setup scripts, reducing the time it takes to configure and deploy your infrastructure.

* `ssh_key` - (Optional) The public SSH key to be added to the server. 

* `ssh_key_name` - (Optional) The name of an existing SSH key to be added to the server. 

* `partitions_config` - (Optional) Configuration for disk partitions.
  * `soft_raid_config` - (Optional) Configuration for software RAID.
    * `name` - (Required) Name of the RAID array.
    * `level` - (Required) RAID level.
    * `disk_type` - (Required) Type of disks to use in the RAID.
  * `disk_partitions` - (Optional) List of disk partitions.
    * `mount` - (Required) Mount point for the partition.
    * `size` - (Optional) Size of the partition in GB. Use only size or size_percent.
    * `size_percent` - (Optional) Size of the partition in percent. Use only size or size_percent.
    * `raid` - (Required) The RAID array name to create the partition on.
    * `fs_type` - (Optional) Filesystem type for the partition.

* `public_subnet_id` - (Optional) ID of the public subnet to connect the server to. 

* `private_subnet` - (Optional) Private subnet to connect the server to. 

* `os_host_name` - (Optional) Hostname for the server.

* `force_update_additional_params` - (Optional) Enables update for additional os params (os_password, user_data, ssh_key, ssh_key_name, partitions_config, os_host_name) without changing os_id. NOTE: installing new os will delete all data on the server.

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
export INFRA_REGION=<selectel_pool>
terraform import selectel_dedicated_servers_server_v1.server_1 <server_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/dbaas), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<server_id>` — Unique identifier of the server.
