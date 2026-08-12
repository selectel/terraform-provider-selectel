---
layout: "selectel"
page_title: "Selectel: selectel_vpc_network_v2"
sidebar_current: "docs-selectel-resource-vpc-network-v2"
description: |-
  Creates and manages a private network in the Selectel cloud using VPC API v2.0.
---

# selectel\_vpc\_network\_v2

Creates and manages a private network using VPC API v2.0. For more information about networks, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/servers/networks/about-networks/).

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

* `admin_state_up` - (Optional) Administrative state of the network. If `false`, the network does not forward packets. If you do not specify the value, the API assigns one.

* `dns_domain` - (Optional) DNS domain of the network. Learn more about [Private DNS](https://docs.selectel.ru/en/cloud-servers/private-dns/).

* `availability_zone_hints` - (Optional) Availability zones that host the network services. Changing this creates a new network.

* `tags` - (Optional) Tags of the network that Terraform manages. Tags added outside Terraform are reported in `all_tags` and are not removed by this resource.

## Attributes Reference

* `status` - Status of the network: `ACTIVE`, `DOWN`, `BUILD`, or `ERROR`. Creating the network waits until the status leaves the transient `BUILD` value.

* `all_tags` - All tags of the network, including the ones added outside Terraform.

* `shared` - Specifies whether the network can be accessed by any project.

* `external` - Specifies whether the network has an external routing facility.

* `mtu` - Maximum transmission unit of the network.

* `port_security_enabled` - Specifies whether port security is enabled by default for the ports of the network.

* `qos_policy_id` - Unique identifier of the QoS policy associated with the network.

* `provider_network_type` - Type of the physical network the network is implemented on, for example, `vxlan`.

* `provider_physical_network` - Physical network the network is implemented on.

* `provider_segmentation_id` - Identifier of the isolated segment on the physical network.

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

* `<project_id>` — Unique identifier of the associated project. The import ID does not contain the project — it is taken from `INFRA_PROJECT_ID`, and the import fails if the variable is not set.

* `<pool>` — Pool where the network is located, for example, `ru-3`. Taken from `INFRA_REGION`, and the import fails if the variable is not set.

* `<network_id>` — Unique identifier of the network, for example, `4f3b1e00-8e39-4a1e-9c8f-1d2b3c4d5e6f`.
