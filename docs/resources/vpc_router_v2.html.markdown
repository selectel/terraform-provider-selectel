---
layout: "selectel"
page_title: "Selectel: selectel_vpc_router_v2"
sidebar_current: "docs-selectel-resource-vpc-router-v2"
description: |-
  Creates and manages a cloud router for Selectel products using public API v2.
---

# selectel\_vpc\_router\_v2

Creates and manages a cloud router using public API v2. For more information about cloud routers, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/cloud-routers/).

The router routes traffic between the subnets attached to it and, when an external gateway is set, to and from an external network. Subnets and ports are attached with [selectel_vpc_router_interface_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_interface_v2), and static routes are managed with [selectel_vpc_router_route_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_route_v2) — this resource does not carry them.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](/docs/providers/selectel/guides/upgrading_to_version_9.html) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_router_v2" "router_1" {
  project_id          = selectel_vpc_project_v2.project_1.id
  region              = "ru-3"
  name                = "router-1"
  description         = "Managed by Terraform"
  tags                = ["production", "networking"]
  external_network_id = data.selectel_vpc_network_v2.external.id
  enable_snat         = true
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new router. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the router is created, for example, `ru-3`. Changing this creates a new router. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `name` - (Optional) Name of the router.

* `description` - (Optional) Description of the router.

* `admin_state_up` - (Optional) Administrative state of the router. The default value is `true`.

* `external_network_id` - (Optional) Unique identifier of the external network the gateway is attached to. Can be changed without replacement, which moves the gateway; removing it detaches the gateway.

* `enable_snat` - (Optional) Specifies whether source NAT is enabled on the gateway. Requires `external_network_id` in the same configuration — a value without it is rejected at plan. Can be changed without replacement.

* `availability_zone_hints` - (Optional) Pool segments the router is requested in. Changing this creates a new router.

* `tags` - (Optional) Tags of the router that are managed by Terraform.

## Attributes Reference

* `status` - Status of the router, for example, `ACTIVE` or `DOWN`.

* `external_fixed_ip` - Gateway addresses assigned by the API:

  * `subnet_id` - Unique identifier of the subnet the address comes from.

  * `ip_address` - Assigned address.

* `availability_zones` - Pool segments the router actually runs in.

* `all_tags` - All tags of the router, including those assigned outside Terraform.

* `blocked` - Specifies whether the router is administratively blocked.

## Import

You can import a router:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_router_v2.router_1 <router_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/). The import ID does not contain the project, so `INFRA_PROJECT_ID` must be set.

* `<pool>` — Pool where the router is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/). The import ID does not contain the pool, so `INFRA_REGION` must be set.

* `<router_id>` — Unique identifier of the router, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the router ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Cloud routers** tab ⟶ copy the ID of the required router. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud/servers/tools/openstack/) command `openstack router list` and copy the `ID` field.

## Behavior notes

A router with attached interfaces is not deleted: the API refuses it with a conflict, and the resource stays in the state — this resource never detaches interfaces on your behalf. Remove the `selectel_vpc_router_interface_v2` resources first.

A router deleted outside Terraform leaves the state on the next refresh without an error, and the next plan re-creates it. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead.
