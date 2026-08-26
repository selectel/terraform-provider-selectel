---
layout: "selectel"
page_title: "Selectel: selectel_vpc_network_v2"
sidebar_current: "docs-selectel-datasource-vpc-network-v2"
description: |-
  Provides information about a private network in Selectel products.
---

# selectel\_vpc\_network\_v2

Provides information about a private network. For more information about private networks, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/).

The data source returns exactly one network. You can combine any of the optional selectors below, and the network must match all of them at the same time. If no network matches the selectors, or if more than one does, the read fails with a diagnostic error instead of picking one arbitrarily.

~> **Note:** If you are upgrading to version 9.0.0 or later, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9) to replace the corresponding OpenStack data source with the Selectel provider data source.

## Example Usage

```hcl
data "selectel_vpc_network_v2" "network_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name       = "network-1"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `region` - (Required) Pool where the network is located, for example, `ru-6`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/infrastructure/availability-matrix/).

* `network_id` - (Optional) Unique identifier of the network. To get the network ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ copy the ID of the required network. As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack network list` and copy the `ID` field.

* `name` - (Optional) Name of the network.

* `description` - (Optional) Description of the network.

* `status` - (Optional) Status of the network: `ACTIVE`, `DOWN`, `BUILD`, or `ERROR`.

* `external` - (Optional) Specifies whether the network has an external routing facility.

* `mtu` - (Optional) Maximum transmission unit of the network.

* `tags` - (Optional) Tags that the network must have.

* `matching_subnet_cidr` - (Optional) CIDR of a subnet attached to the network. Selects the network that owns a subnet with this CIDR.

## Attributes Reference

* `admin_state_up` - Administrative state of the network.

* `dns_domain` - DNS domain of the network. Learn more about [Private DNS](https://docs.selectel.ru/en/cloud-servers/private-dns/).

* `availability_zone_hints` - Pool segments requested for hosting the network services.

* `availability_zones` - Pool segments where the network services are actually hosted.

* `subnets` - Unique identifiers of the subnets attached to the network.

* `all_tags` - All tags of the network.

* `shared` - Specifies whether the network can be accessed by any project.

* `port_security_enabled` - Specifies whether port security is enabled by default for the ports of the network.

* `blocked` - Specifies whether the network is administratively blocked.

* `is_public` - Specifies whether the network is public.

* `is_dns_enabled` - Specifies whether the Private DNS service is enabled for the network. Learn more about [Private DNS](https://docs.selectel.ru/en/cloud-servers/private-dns/).
