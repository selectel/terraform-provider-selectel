---
layout: "selectel"
page_title: "Selectel: selectel_vpc_subnet_v2"
sidebar_current: "docs-selectel-resource-vpc-subnet-v2"
description: |-
  Creates and manages a public subnet for Selectel products using public API v2.
---

# selectel\_vpc\_subnet_v2

Creates and manages a public subnet using public API v2. For more information about public subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/servers/networks/about-networks/).

For private networks and subnets, use [openstack\_networking\_network\_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_network_v2) and [openstack\_networking\_subnet\_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_subnet_v2) resources of the OpenStack provider.

## Example Usage

```hcl
resource "selectel_vpc_subnet_v2" "subnet_1" {
  project_id    = selectel_vpc_project_v2.project_1.id
  region        = "ru-3"
  ip_version    = "ipv4"
  prefix_length = 29
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new public subnet. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the public subnet is located, for example, `ru-3`. Changing this creates a new public subnet. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `ip_version` - (Optional) Internet protocol version supported in the public subnet. The only available value is `ipv4`.

* `prefix_length` - (Optional) Prefix length of the public subnet. The default value is `29`. Changing this creates a new public subnet.

## Attributes Reference

* `cidr` - CIDR of the public subnet.

* `network_id` - Unique identifier of the associated OpenStack network. Learn more about the [openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_network_v2) resource in the official OpenStack documentation.

* `subnet_id` - Unique identifier of the associated OpenStack subnet. Learn more about the [openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_subnet_v2) resource in the official OpenStack documentation.

* `status` - Status of the public subnet.

* `servers` - List of the cloud servers that are located in the public subnet.

  * `id` - Unique identifier of the cloud server.

  * `name` - Name of the cloud server.

  * `status` - Status of the cloud server.

## Import

You can import a public subnet:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_vpc_subnet_v2.subnet_1 <public_subnet_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<public_subnet_id>` is a unique identifier of the public subnet, for example, `2060`. To get the public subnet ID, use [Selectel Cloud Management API](https://developers.selectel.ru/docs/selectel-cloud-platform/main-services/selectel_cloud_management_api/).
