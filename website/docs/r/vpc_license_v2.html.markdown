---
layout: "selectel"
page_title: "Selectel: selectel_vpc_license_v2"
sidebar_current: "docs-selectel-resource-vpc-license-v2"
description: |-
  Manages a license for Selectel cloud servers using public API v2.
---

# selectel\_vpc\_license_v2

Manages a license for cloud servers using public API v2.

## Example Usage

```hcl
resource "selectel_vpc_license_v2" "license_windows_2016_standard" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-2"
  type       = "license_windows_2012_standard"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new license. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where you can use the license, for example, `ru-3`. The cloud server must be located in the pool. Changing this creates a new license. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `type` - (Required) Type of the license. Changing this creates a new license. Available values are `license_windows_2012_standard`, `license_windows_2016_standard`, `license_windows_2019_standard`.

## Attributes Reference

* `status` - License status.

* `servers` - Cloud servers that use the license.

  * `id` - Unique identifier of the cloud server.

  * `name` - Name of the cloud server.

  * `status` - Status of the cloud server.

* `network_id` - Unique identifier of the associated OpenStack network. Learn more about the [openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_network_v2) resource in the official OpenStack documentation.

* `subnet_id` - Unique identifier of the associated OpenStack subnet. Learn more about the [openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_subnet_v2) resource in the official OpenStack documentation.

* `port_id` - Unique identifier of the associated OpenStack port. Learn more about the [openstack_networking_port_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_port_v2) resource in the official OpenStack documentation.

## Import

You can import a license:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_vpc_license_v2.license_1 <license_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<license_id>` — Unique identifier of the license, for example, `4123`. To get the license ID, use [Selectel Cloud Management API](https://developers.selectel.ru/docs/selectel-cloud-platform/main-services/selectel_cloud_management_api/).


