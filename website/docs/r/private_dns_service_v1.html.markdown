---
layout: "selectel"
page_title: "Selectel: selectel_private_dns_service_v1"
sidebar_current: "docs-selectel-private-dns-service-v1"
description: |-
  Creates and manages a DNS service in Selectel Private DNS using public API v1
---

# selectel\_private\_dns\_service\_v1

Creates and manages a private DNS service for connecting a network to private DNS using public API v1. For more information about private DNS, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/private-dns/).

## Example usage

```hcl
resource "selectel_private_dns_service_v1" "service_1" {
	region     = "ru-1"
	project_id = selectel_vpc_project_v2.project_1.id
	network_id = openstack_networking_network_v2.network_1.id

	depends_on = [
		openstack_networking_network_v2.network_1
    ]
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the network is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `network_id` - (Required) Unique identifier of a network to connect to the DNS service. Retrieved from the [openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_network_v2) resource. Learn more about [Networks](https://docs.selectel.ru/en/cloud-networks/private-networks-and-subnets/).

## Attributes Reference

* `addresses` - List of the DNS service IP addresses:
    * `address` - IP addresses in a subnet for accessing the DNS service.
    * `cidr` - Subnet IP address range in CIDR notation.
