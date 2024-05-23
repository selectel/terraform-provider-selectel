---
layout: "selectel"
page_title: "Selectel: selectel_vpc_project_v2"
sidebar_current: "docs-selectel-resource-vpc-project-v2"
description: |-
  Creates and manages a Selectel project using public API v2.
---

# selectel\_vpc\_project_v2

Creates and manages a project using public API v2. For more information about projects, see the [official Selectel documentation](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

## Example Usage

### Project with quotas

```hcl
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
```

### Project with external panel

```hcl
resource "selectel_vpc_project_v2" "project_1" {
  name       = "project_1"
  custom_url = "project-123.selvpc.ru"
  theme = {
    color = "2753E9"
  }
}
```

## Argument Reference

* `name` - (Required) Project name.

* `quotas` - (Optional) Array of quotas for the project. Learn more about [Project limits and quotas](https://docs.selectel.ru/en/control-panel-actions/projects/quotas/).

  * `resource_name` - (Required) Resource name. To get the name of the resource, use [Selectel Cloud Quota Management API](https://developers.selectel.ru/docs/selectel-cloud-platform/main-services/cloud-quota-management/).

  * `resource_quotas` - (Required) Array of quotas for the resource.

    * `region` - (Optional) Pool where the resource is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

    * `zone` - (Optional) Pool segment where the resource is located, for example, `ru-3a`. Learn more about available pool segments in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

    * `value` - (Required) Quota value. The value cannot exceed the project limit. To get the project limit, in the [Control panel](https://my.selectel.ru/vpc/quotas/), go to **Cloud Platform** ⟶ **Quotas**. The project limit for the resource is in the **Quota** column. Learn more about [Project limits and quotas](https://docs.selectel.ru/en/control-panel-actions/projects/quotas/).

* `custom_url` - (Optional) URL of the project in the external panel. The available value is the third-level domain, for example, `123456.selvpc.ru` or `project.example.com`. Learn more [how to set up access to external panel](https://docs.selectel.ru/en/control-panel-actions/account/external-panel/).

* `theme` - (Optional) Additional theme settings for the external panel.

  * `color` - (Optional) Fill color of the toolbar in hex format.

  * `logo` - (Optional) URL of the logo on the toolbar.

## Attributes Reference

* `url` - Project URL. Created automatically and you cannot change it.

* `enabled` - Project status. Possible values are `active` and `disabled`.

* `all_quotas` - List of quotas. Can differ from the values that are set in the `quotas` block, if all available quotas for the project are automatically applied.

## Import

You can import a project:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_vpc_project_v2.project_1 <project_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the project, for example, `a07abc12310546f1b9291ab3013a7d75`. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project.

