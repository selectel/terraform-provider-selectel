---
layout: "selectel"
page_title: "Selectel: selectel_vpc_public_port_v1"
sidebar_current: "docs-selectel-resource-vpc-public-port-v1"
description: |-
  Creates and manages a direct public IP address (public port) for Selectel products using public API v1
---

# selectel\_vpc\_public_port\_v1

Creates and manages a direct public IP address (their public port) using public API v1. For more information about direct public IP address, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/direct-public-ip-addresses).

## Example usage

```hcl
resource "selectel_vpc_public_port_v1" "port_1" {
  project_id         = selectel_vpc_project_v2.project_1.id
  region             = "ru-3"
  description        = "my-own-direct-public-ip-port"
  admin_state_up     = false
  security_group_ids = [openstack_networking_secgroup_v2.sg_1.id]
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier (no-hyphens) of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Region where the resource should be created, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `description` - (Optional) Resource description.

* `admin_state_up` - (Optional) Specifies whether the resource is administratively up (`true`) or down (`false`). Defaults to up (`true`).

* `security_group_ids` - (Optional) List of security group identifiers to associate with the resource. If omitted, the resource is automatically assigned the default security group of the project.

## Attributes Reference

* `network_id` - Identifier of the network where the resource is allocated

* `ip_address` - Direct public IP address assigned to the resource

* `subnet` - CIDR of the subnet where the resource is allocated

* `gateway` - IP address of the network gateway for the resource
