---
layout: "selectel"
page_title: "Selectel: selectel_vpc_network_v2"
sidebar_current: "docs-selectel-datasource-vpc-network-v2"
description: |-
  Provides information about a private network in the Selectel cloud using Neutron API v2.0.
---

# selectel\_networking\_network\_v2

Provides information about a private network. For more information about networks, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/servers/networks/about-networks/).

The data source returns exactly one network. If no network matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

## Example Usage

```hcl
data "selectel_vpc_network_v2" "network_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "network-1"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the network is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

You can combine any of the following selectors. The network must match all of them at the same time:

* `network_id` - (Optional) Unique identifier of the network.

* `name` - (Optional) Name of the network.

* `description` - (Optional) Description of the network.

* `status` - (Optional) Status of the network: `ACTIVE`, `DOWN`, `BUILD`, or `ERROR`.

* `external` - (Optional) Specifies whether the network has an external routing facility.

* `mtu` - (Optional) Maximum transmission unit of the network.

* `tags` - (Optional) Tags that the network must have.

* `matching_subnet_cidr` - (Optional) CIDR of a subnet attached to the network. Selects the network that owns a subnet with this CIDR.

## Attributes Reference

* `admin_state_up` - Administrative state of the network.

* `dns_domain` - DNS domain of the network.

* `availability_zone_hints` - Availability zones that host the network services.

* `subnets` - Unique identifiers of the subnets attached to the network.

* `all_tags` - All tags of the network.

* `shared` - Specifies whether the network can be accessed by any project.

* `port_security_enabled` - Specifies whether port security is enabled by default for the ports of the network.

* `qos_policy_id` - Unique identifier of the QoS policy associated with the network.

* `provider_network_type` - Type of the physical network the network is implemented on.

* `provider_physical_network` - Physical network the network is implemented on.

* `provider_segmentation_id` - Identifier of the isolated segment on the physical network.

* `blocked` - Specifies whether the network is administratively blocked.

* `is_public` - Specifies whether the network is public.

* `is_dns_enabled` - Specifies whether the DNS service is enabled for the network.

~> **Note:** The returned attributes report the actual state of the network and cannot be used to change it. The attributes the cloud's Neutron policy reserves for the Keystone admin role are returned here but are not available as selectors.
