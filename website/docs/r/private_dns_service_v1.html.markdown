---
layout: "selectel"
page_title: "Selectel: selectel_private_dns_service_v1"
sidebar_current: "docs-selectel-private-dns_service-v1"
description: |-
  Creates and manages a dns service in Selectel Private DNS using public API v1.
---

# selectel\_private\_dns\_service\_v1

Creates and manages a private dns service using public API v1. For more information about private dns services, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/private-dns/service).

## Example usage

```hcl
resource "selectel_private_dns_service_v1" "service" {
    region = "ru-1"
    project_id = selectel_vpc_project_v2.project_1.id
    network_id = openstack_networking_network_v2.network_one.id

    depends_on = [
		  openstack_networking_network_v2.network_one
    ]
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `region` - (Required) Pool where the backup plan is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).

* `network_id` - (Required) The ID of the network in which the DNS resolver will be created. Retrieved from the [openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_network_v2). Learn more about [Networks](https://docs.selectel.ru/en/cloud-networks/private-networks-and-subnets/).

## Attributes Reference

* `addresses` - List of service ip addresses.
  * `address` - IP address.
  * `cidr` - IP subnet CIDR.