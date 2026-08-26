---
layout: "selectel"
page_title: "Selectel: selectel_vpc_port_forwarding_v2"
sidebar_current: "docs-selectel-resource-vpc-port-forwarding-v2"
description: |-
  Creates and manages a port-forwarding rule of a public floating IP address for Selectel products using public API v2.
---

# selectel\_vpc\_port\_forwarding\_v2

Creates and manages a port-forwarding rule of a public floating IP address using public API v2. For more information about public floating IP addresses, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/public-floating-ip-addresses/).

Traffic that arrives at an external port of the public floating IP address is forwarded to an internal address and port. The public floating IP address is managed with the [selectel_vpc_floatingip_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_floatingip_v2) resource.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to migrate your resources from the OpenStack Terraform provider to the Selectel provider.

## Example Usage

```hcl
resource "selectel_vpc_port_forwarding_v2" "port_forwarding_1" {
  project_id          = selectel_vpc_project_v2.project_1.id
  region              = "ru-6"
  floatingip_id       = selectel_vpc_floatingip_v2.floatingip_1.id
  internal_port_id    = selectel_vpc_port_v2.port_1.id
  internal_ip_address = "192.168.199.50"
  internal_port       = 80
  external_port       = 8080
  protocol            = "tcp"
  depends_on          = [selectel_vpc_router_interface_v2.interface_1]
}
```

~> **Note:** The subnet of the internal port must already have a route to the external network the public floating IP address comes from. No argument of this resource refers to the router interface that provides the route, so declare the dependency with `depends_on`.

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Changing this creates a new rule. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the public floating IP address is located, for example, `ru-6`. Changing this creates a new rule. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `floatingip_id` - (Required) Unique identifier of the public floating IP address the rule belongs to. Changing this creates a new rule, because the rule exists under one public floating IP address. Retrieved from the [selectel_vpc_floatingip_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_floatingip_v2) resource.

* `internal_port_id` - (Required) Unique identifier of the port traffic is forwarded to. Can be changed without replacement. Retrieved from the [selectel_vpc_port_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2) resource.

* `internal_ip_address` - (Required) Address traffic is forwarded to, which must be one of the fixed IPs of `internal_port_id`. Can be changed without replacement.

* `internal_port` - (Required) Port traffic is forwarded to. Can be changed without replacement. The available range is from 1 to 65535.

* `external_port` - (Required) Port of the public floating IP address that traffic arrives at. Can be changed without replacement. The available range is from 1 to 65535.

* `protocol` - (Required) Protocol of the forwarded traffic. Available values are `tcp`, `udp`, `icmp`, `ipv6-icmp`, `sctp`, and `dccp`.

* `description` - (Optional) Description of the rule.

## Attributes Reference

No attributes are exported beyond the arguments. The identifier of the resource is the identifier the API assigns to the rule.

## Import

You can import a rule by the identifier of its public floating IP address and the identifier of the rule:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<pool>
terraform import selectel_vpc_port_forwarding_v2.pf_1 <floatingip_id>/<id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/user-types).

* `<password>` — Password of the service user.

* `<project_id>` — Unique identifier of the associated project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `<pool>` — Pool where the public floating IP address is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `<floatingip_id>` — Unique identifier of the public floating IP address, for example, `1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809`. To get the public floating IP address ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Public Floating IP addresses** tab ⟶ copy the ID of the required address. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack floating ip list` and copy the `ID` field.

* `<id>` — Unique identifier of the rule, for example, `9f8e7d6c-5b4a-3928-1706-f5e4d3c2b1a0`. To get the rule ID, use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack floating ip port forwarding list <floatingip_id>` and copy the `ID` field. The Control panel does not show the rule ID.

## Behavior notes

A rule removed outside Terraform, or one whose public floating IP address was deleted, leaves the state on the next refresh without an error, and the next plan re-creates it — the rule cannot exist without its public floating IP address. A denied access or a transport failure keeps the resource in the state and reports a diagnostic error instead.
