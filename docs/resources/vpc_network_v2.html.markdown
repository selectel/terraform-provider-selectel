---
layout: "selectel"
page_title: "Selectel: selectel_vpc_network_v2"
sidebar_current: "docs-selectel-resource-vpc-network-v2"
description: |-
  Creates and manages a private network for Selectel products using public API v2.
---

# selectel\_vpc\_network\_v2

Creates and manages a private network using public API v2. For more information about private networks, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/).

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_network_v2" "network_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-3"
  name        = "network-1"
  description = "Managed by Terraform"
  tags        = ["production", "networking"]
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new network. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the network is located, for example, `ru-3`. Changing this creates a new network. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `name` - (Optional) Name of the network. If you do not specify a name, the API assigns one.

* `description` - (Optional) Description of the network.

* `admin_state_up` - (Optional) Administrative state of the network. If `false`, the network does not forward packets. The default value is `true`.

* `dns_domain` - (Optional) DNS domain of the network. Learn more about [Private DNS](https://docs.selectel.ru/en/cloud-servers/private-dns/).

* `availability_zone_hints` - (Optional) Pool segments requested for hosting the network services. Changing this creates a new network.

* `tags` - (Optional) Tags of the network that Terraform manages. Tags added outside Terraform are reported in `all_tags` and are not removed by this resource.

## Attributes Reference

* `status` - Status of the network: `ACTIVE`, `DOWN`, `BUILD`, or `ERROR`.

* `subnets` - Unique identifiers of the subnets attached to the network.

* `availability_zones` - Pool segments where the network services are actually hosted. The value is assigned by the API and can differ from the requested `availability_zone_hints`.

* `all_tags` - All tags of the network, including the ones added outside Terraform.

* `shared` - Specifies whether the network can be accessed by any project.

* `external` - Specifies whether the network has an external routing facility.

* `mtu` - Maximum transmission unit of the network.

* `port_security_enabled` - Specifies whether port security is enabled by default for the ports of the network.

* `blocked` - Specifies whether the network is administratively blocked.

* `is_public` - Specifies whether the network is public.

* `is_dns_enabled` - Specifies whether the Private DNS service is enabled for the network. Learn more about [Private DNS](https://docs.selectel.ru/en/cloud-servers/private-dns/).

## Import

You can import a network:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_network_v2.network_1 <network_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). The import ID does not contain the project, so `INFRA_PROJECT_ID` must be set.

* `<pool>` — Pool where the network is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/). The import ID does not contain the pool, so `INFRA_REGION` must be set.

* `<network_id>` — Unique identifier of the network, for example, `4f3b1e00-8e39-4a1e-9c8f-1d2b3c4d5e6f`. To get the network ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ copy the ID of the required network. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack network list` and copy the `ID` field.
