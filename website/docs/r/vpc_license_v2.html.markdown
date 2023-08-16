---
layout: "selectel"
page_title: "Selectel: selectel_vpc_license_v2"
sidebar_current: "docs-selectel-resource-vpc-license-v2"
description: |-
  Manages a V2 license resource within Selectel VPC.
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

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new license. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Required) Pool where you can use the license, for example, `ru-3`. The cloud server must be located in the pool. Changing this creates a new license. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/).

* `type` - (Required) Type of the license. Changing this creates a new license. Available values are `license_windows_2012_standard`, `license_windows_2016_standard`, `license_windows_2019_standard`.

## Attributes Reference

* `status` - License status.

* `servers` - Cloud servers that use the license.

  * `id` - Unique identifier of the cloud server.

  * `name` - Name of the cloud server.

  * `status` - Status of the cloud server.

* `network_id` - Unique identifier of the associated OpenStack network. Learn more about the [openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_network_v2) resource in the official OpenStack documentation.

* `subnet_id` - Unique identifier of the associated OpenStack subnet. Learn more about the [openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_subnet_v2) resource in the official OpenStack documentation.

* `port_id` - Unique identifier of the associated OpenStack port. Learn more about the [openstack_networking_port_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_port_v2) resource in the official OpenStack documentation.

## Import

You can import a license:

```shell
terraform import selectel_vpc_license_v2.license_1 <license_id>
```

where `<license_id>` is a unique identifier of the license, for example, `4123`. To get the license ID, use [Selectel Cloud Management API](https://developers.selectel.ru/docs/selectel-cloud-platform/main-services/selectel_cloud_management_api/).

### Environment Variables

For import, you must set the environment variable `SEL_TOKEN=<selectel_api_token>`,

where `<selectel_api_token>` is a Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).
