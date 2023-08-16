---
layout: "selectel"
page_title: "Selectel: selectel_vpc_subnet_v2"
sidebar_current: "docs-selectel-resource-vpc-subnet-v2"
description: |-
  Manages a V2 subnet resource within Selectel VPC.
---

# selectel\_vpc\_subnet_v2

Creates and manages a public subnet using public API v2. For more information about public subnets, see the [official Selectel documentation](https://docs.selectel.ru/cloud/servers/networks/about-networks/).

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

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new public subnet. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Required) Pool where the public subnet is located, for example, `ru-3`. Changing this creates a new public subnet. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/).

* `ip_version` - (Optional) Internet protocol version supported in the public subnet. The only available value is `ipv4`.

* `prefix_length` - (Optional) Prefix length of the public subnet. The default value is `29`. Changing this creates a new public subnet.

## Attributes Reference

* `cidr` - CIDR of the public subnet.

* `network_id` - Unique identifier of the associated OpenStack network. Learn more about the [openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_network_v2) resource in the official OpenStack documentation.

* `subnet_id` - Unique identifier of the associated OpenStack subnet. Learn more about the [openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/data-sources/networking_subnet_v2) resource in the official OpenStack documentation.

* `status` - Status of the public subnet.

* `servers` - List of the cloud servers that are located in the public subnet.

  * `id` - Unique identifier of the cloud server.

  * `name` - Name of the cloud server.

  * `status` - Status of the cloud server.

## Import

You can import a public subnet:

```shell
terraform import selectel_vpc_subnet_v2.subnet_1 <public_subnet_id>
```

where `<public_subnet_id>` is a unique identifier of the public subnet, for example, `2060`. To get the public subnet ID, use [Selectel Cloud Management API](https://developers.selectel.ru/docs/selectel-cloud-platform/main-services/selectel_cloud_management_api/).

### Environment Variables

For import, you must set the environment variable `SEL_TOKEN=<selectel_api_token>`,

where `<selectel_api_token>` is a Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).
