---
page_title: "Selectel: selectel_vpc_subnet_v2"
description: |-
  Creates and manages a public subnet for Selectel products using the v2 public API.
---

# selectel\_vpc\_subnet\_v2

!> **Deprecated since 9.1.0**
This resource is no longer maintained and will be removed in a future major release. To continue receiving updates, please migrate to [selectel\_vpc\_public\_subnet\_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_public_subnet_v2). To migrate an existing resource, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9_1).

Creates and manages a public subnet using the v2 public API. For more information about public subnets, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/public-subnets/).

For project networks and their subnets, use [selectel\_vpc\_network\_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_v2) and [selectel\_vpc\_network\_subnet\_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2).

Creates a public subnet in a service network. You cannot select a network during subnet creation or modify service network settings.

## Example Usage

```terraform
resource "selectel_vpc_subnet_v2" "subnet_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new public subnet. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the public subnet is located, for example, `ru-6`. Changing this creates a new public subnet. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `ip_version` - (Optional) Internet protocol version supported in the public subnet. Changing this creates a new public subnet. The only available value is `ipv4`, which is also the default value.

* `prefix_length` - (Optional) Prefix length of the public subnet. Available values are from `24` to `29`. The default value is `29`. Changing this creates a new public subnet.

## Attributes Reference

* `cidr` - CIDR of the public subnet.

* `network_id` - Unique identifier of the parent service network — an identifier of the same type as in the [selectel_vpc_network_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_v2) resource.

* `subnet_id` - Unique identifier of the associated subnet in the parent service network — an identifier of the same type as in the [selectel_vpc_network_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) resource. Do not use it to import the public subnet — the import uses a different identifier, see [Import](#import).

* `status` - Status of the public subnet.

* `servers` - List of cloud servers that are located in the public subnet.

  * `id` - Unique identifier of the cloud server.

  * `name` - Name of the cloud server.

  * `status` - Status of the cloud server.

## Import

You can import a public subnet:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
terraform import selectel_vpc_subnet_v2.subnet_1 <public_subnet_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/user-types).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `<public_subnet_id>` — Unique identifier of the public subnet — the `id` field in the response of the [Get list of subnets](https://docs.selectel.ru/en/api/cloud-projects-and-resources/#tag/Subnets/operation/listSubnets) method in [Cloud Platform Projects and Resources API](https://docs.selectel.ru/en/api/cloud-projects-and-resources/). Do not use the value of the `subnet_id` attribute — it identifies the associated subnet in the parent service network, not the public subnet.
