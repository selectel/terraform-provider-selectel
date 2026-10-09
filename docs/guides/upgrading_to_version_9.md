---
page_title: "Upgrading Terraform Selectel Provider to version 9.0.0"
description: |-
  How to upgrade Terraform Selectel Provider to version 9.0.0.
---

# Upgrading Terraform Selectel Provider to version 9.0.0

In version 9.0.0, Terraform Selectel Provider introduces the following changes:

- deprecates the `selectel_vpc_subnet_v2` resource and adds the `selectel_vpc_public_subnet_v2` resource as a replacement
- adds networking resources and data sources that replace the corresponding `terraform-provider-openstack` resources and data sources. For the exact list, see [Selectel and OpenStack resource and data source mapping](#selectel-openstack-resource-data-source-mapping).

You can continue to use the `terraform-provider-openstack` resources, but we recommend migrating to the Selectel resources, because their behavior matches the actual Selectel API.

This guide explains how to migrate without recreating the resources. It covers:

- [Upgrading the version](#upgrade-the-version)
- [Environment variables for import](#environment-variables-for-import)
- [Migrating the deprecated `selectel_vpc_subnet_v2` resource](#migrating-the-deprecated-selectel_vpc_subnet_v2-resource)
- [Migrating resources and data sources from terraform-provider-openstack](#migrating-resources-and-data-sources-from-terraform-provider-openstack).

Before upgrading to version 9.0.0, upgrade to the most recent 8.X version of the provider and ensure that your environment successfully runs `terraform plan`. You should not see changes you do not expect or deprecation notices.

## Upgrade the version {#upgrade-the-version}

Upgrade the version, before any of the migrations below.

1. In the Terraform configuration, update the version constraints:

   ```hcl
   terraform {
     required_providers {
       selectel = {
         source  = "selectel/selectel"
         version = "~> 9.0"
       }
       openstack = {
         source  = "terraform-provider-openstack/openstack"
         version = "3.4.0"
       }
     }
   }
   ```

2. To download the new version, initialize the Terraform configuration:

   ```bash
   terraform init -upgrade
   ```

3. Back up the `.tfstate` file.

## Migrating the deprecated `selectel_vpc_subnet_v2` resource

The `selectel_vpc_subnet_v2` resource is deprecated. Existing configurations continue to work, but use `selectel_vpc_public_subnet_v2` for new configurations. Both resources manage the same kind of public subnet and use the same arguments and attributes.

Before you start, [upgrade the version](#upgrade-the-version).

1. Copy the public subnet ID:

   ```bash
   terraform state show selectel_vpc_subnet_v2.example
   ```

2. In the configuration files, change the resource type to `selectel_vpc_public_subnet_v2` and update every reference to it. Keep the resource name and arguments unchanged. For example:

   ```hcl
   resource "selectel_vpc_public_subnet_v2" "example" {
     project_id    = selectel_vpc_project_v2.project_1.id
     region        = "ru-3"
     ip_version    = "ipv4"
     prefix_length = 29
   }
   ```

3. Remove the old resource address from the `.tfstate` file:

   ```bash
   terraform state rm selectel_vpc_subnet_v2.example
   ```

4. Import the same public subnet under the new resource type. Set the [environment variables](#environment-variables-for-import) first:

   ```bash
   export OS_DOMAIN_NAME=<account_id>
   export OS_USERNAME=<username>
   export OS_PASSWORD=<password>
   export INFRA_PROJECT_ID=<project_id>
   terraform import selectel_vpc_public_subnet_v2.example <public_subnet_id>
   ```

   where `<public_subnet_id>` is the ID that you copied in step 1.

5. To ensure that Terraform does not plan to create or destroy the public subnet, preview the changes:

   ```bash
   terraform plan
   ```

6. Apply the changes:

   ```bash
   terraform apply
   ```

## Migrating resources and data sources from terraform-provider-openstack

Version 9.0.0 ships no migration tool: no generator for `terraform state rm` and `terraform import` commands, no bulk-import script, and no conversion utility for `.tf` files.

Migration is manual and per-resource. If you manage dozens of resources, migrate each of them separately and check every imported object with `terraform plan` before applying anything.

### Selectel and OpenStack resource and data source mapping {#selectel-openstack-resource-data-source-mapping}

Each row of the first table maps one `terraform-provider-openstack` resource to the resource that replaces it.

| terraform-provider-openstack | terraform-provider-selectel |
| ---------------------------- | --------------------------- |
| `openstack_networking_network_v2` | `selectel_vpc_network_v2` |
| `openstack_networking_subnet_v2` | `selectel_vpc_network_subnet_v2` |
| `openstack_networking_subnet_route_v2` | `selectel_vpc_network_subnet_route_v2` |
| `openstack_networking_port_v2` | `selectel_vpc_port_v2` |
| `openstack_networking_portforwarding_v2` | `selectel_vpc_port_forwarding_v2` |
| `openstack_networking_router_v2` | `selectel_vpc_router_v2` |
| `openstack_networking_router_interface_v2` | `selectel_vpc_router_interface_v2` |
| `openstack_networking_router_route_v2` | `selectel_vpc_router_route_v2` |
| `openstack_networking_router_routes_v2` | `selectel_vpc_router_routes_v2` |
| `openstack_networking_floatingip_associate_v2` | `selectel_vpc_floatingip_associate_v2` |
| `openstack_networking_secgroup_v2` | `selectel_vpc_secgroup_v2` |
| `openstack_networking_secgroup_rule_v2` | `selectel_vpc_secgroup_rule_v2` |
| `openstack_networking_port_secgroup_associate_v2` | `selectel_vpc_port_secgroup_associate_v2` |
| `openstack_networking_rbac_policy_v2` | `selectel_vpc_rbac_policy_v2` |

Networking resources that are not listed in the table are not supported by the Selectel provider.

Five of the fourteen resources have a corresponding data source. The others either represent links that have no independent object to search for, or, in the case of the floating IP association, keep allocation and deletion in the existing resource. The second table maps the data sources:

| terraform-provider-openstack | terraform-provider-selectel |
| ---------------------------- | --------------------------- |
| `openstack_networking_network_v2` | `selectel_vpc_network_v2` |
| `openstack_networking_subnet_v2` | `selectel_vpc_network_subnet_v2` |
| `openstack_networking_port_v2` | `selectel_vpc_port_v2` |
| `openstack_networking_router_v2` | `selectel_vpc_router_v2` |
| `openstack_networking_secgroup_v2` | `selectel_vpc_secgroup_v2` |

### Schema differences

Read this section before you migrate: it lists what changes in the schemas, so you can prepare the configuration. The migration procedures below refer to it.

#### Attributes that are new and read-only

The Selectel resources have attributes that the OpenStack schema does not have. These attributes report platform state: the provider reflects them but never sends them, so there is nothing to add to your configuration.

| Entity | New read-only attributes |
| ------ | ------------------------ |
| [network](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_v2) | `status`, `subnets`, `availability_zones`, `blocked`, `is_public`, `is_dns_enabled` |
| [subnet](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) | `host_routes`, `blocked` |
| [port](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2) | `status`, `blocked`, `dhcp_blocked` |
| [router](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2) | `status`, `availability_zones`, `blocked` |
| [security group](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_secgroup_v2) | `shared`, `security_group_rules` |

A security group has no `blocked` attribute. Its `shared` attribute reports whether an RBAC policy grants another project access to the group. Manage that access with `selectel_vpc_rbac_policy_v2`, not on the group itself. `security_group_rules` contains the identifiers of the rules that belong to the group.

#### Attributes that are new and configurable

The Selectel resources have attributes that the OpenStack schema does not have.

* In the `selectel_vpc_router_v2` resource, an `ha` argument enables a high availability mode. The `openstack_networking_router_v2` resource has no such argument. The argument is optional: the provider reads the current value from the Selectel API, so an imported router keeps its mode and you do not need to add `ha` to the configuration. Set it only if you want to change the mode. Changing `ha` on an existing router disables the router for the time of the change, so traffic through it is interrupted. For more information, see the [`selectel_vpc_router_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2) resource.

#### Attributes that become read-only or are not exposed

If your OpenStack resource configuration sets any attributes listed below, remove them when changing the resource type. The Selectel resource does not accept them as configuration arguments, so Terraform reports an unsupported-argument error before building a plan. Read-only attributes remain available for references. Attributes listed as not exposed must also be removed from references.

| Entity | Read-only in the Selectel resource | Not exposed at all |
| ------ | ---------------------------------- | ------------------ |
| [network](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_v2) | `shared`, `external`, `mtu`, `port_security_enabled` | `tenant_id`, `segments`, `transparent_vlan`, `value_specs`, `qos_policy_id` |
| [subnet](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) | `service_types` | `tenant_id`, `prefix_length`, `dns_publish_fixed_ip`, `segment_id`, `subnetpool_id`, `value_specs` |
| [port](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2) | `port_security_enabled` | `tenant_id`, `value_specs`, `binding`, `dns_assignment`, `qos_policy_id` |
| [router](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2) | `external_fixed_ip` | `tenant_id`, `external_qos_policy_id`, `distributed`, `external_subnet_ids`, `flavor_id`, `value_specs`, `vendor_options` |
| [security group](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_secgroup_v2) | — | `tenant_id` |

The Selectel resources use the required `project_id` argument instead of `tenant_id`. Remove `tenant_id` rather than copying it alongside `project_id`.

#### Changed defaults

The `admin_state_up` argument defaults to `true` in the following resources:

* [`selectel_vpc_network_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_v2)
* [`selectel_vpc_port_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2)
* [`selectel_vpc_router_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2).

To preserve an imported object that is administratively down, set `admin_state_up = false` explicitly. Otherwise, Terraform plans to enable it.

#### Other schema changes

* [`selectel_vpc_network_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_v2): `availability_zone_hints` changes from an unordered `Set` to an ordered `List`.
* [`selectel_vpc_network_subnet_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2):
  * `service_types` changes from a configurable `List` to a read-only `Set`. Remove it from the managed subnet configuration, and do not use numeric indexes when referencing it.
  * `prefix_length` and subnet-pool allocation are not supported. The OpenStack resource uses `prefix_length` with `subnetpool_id` to allocate a CIDR automatically. The Selectel resource requires an explicit `cidr`. Before migration, copy the allocated CIDR from the existing resource state into the new configuration.
* [`selectel_vpc_port_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2): `all_fixed_ips` changes from a list of IP address strings to a list of objects with `subnet_id` and `ip_address`. For example, replace `all_fixed_ips[0]` with `all_fixed_ips[0].ip_address`.
* [`selectel_vpc_secgroup_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_secgroup_v2): `description` is no longer computed when omitted. Copy a non-empty existing description into the configuration if it must be preserved.
* [`selectel_vpc_rbac_policy_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_rbac_policy_v2): `target_tenant` is renamed to `target_project_id`. It still holds the ID of the project the access is granted to. Rename the argument when changing the resource type.

### Migrating data sources

Data sources do not manage the lifecycle of remote objects, so they do not need `terraform state rm` or `terraform import`.

Before you start, [upgrade the version](#upgrade-the-version).

1. In the configuration files, replace the OpenStack data source with the corresponding Selectel data source from the [mapping table](#selectel-openstack-resource-data-source-mapping) and add the `project_id` and `region` arguments.

2. If you use the `openstack_networking_subnet_v2` data source, apply the [schema changes for the subnet data source](#schema-changes-for-the-subnet-data-source).

3. Update references to the attributes of the data sources.

4. To verify that the new data source selects the expected object and that all references use attributes available in its schema, preview the changes:

   ```bash
   terraform plan
   ```

5. Apply the changes:

   ```bash
   terraform apply
   ```

#### Schema changes for the subnet data source

When migrating `openstack_networking_subnet_v2`, make these schema changes:

| OpenStack data source | Selectel data source | Migration action |
| --------------------- | -------------------- | ---------------- |
| `dhcp_enabled` selector | `enable_dhcp` selector | Rename the argument. `enable_dhcp` is also the computed result. |
| `allocation_pools` attribute | `allocation_pool` attribute | Rename every reference. The Selectel attribute is an unordered set rather than a list. |
| `dns_nameservers` attribute | `dns_nameservers` attribute | The name is unchanged, but the Selectel attribute is an ordered list rather than a set. |

Because `allocation_pool` is a set, do not access it with an index such as `allocation_pools[0]`. Use a `for` expression when processing every pool, or `one(...)` when exactly one pool is expected.

For example:

```hcl
data "selectel_vpc_network_subnet_v2" "subnet_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-3"
  network_id  = selectel_vpc_network_v2.network_1.id
  cidr        = "192.168.199.0/24"
  enable_dhcp = true
}
```

### Migrating managed resources

Changing from `openstack_networking_*` to `selectel_vpc_*` changes the Terraform resource type. `terraform state mv` cannot do that: its source and destination must have the same resource type. Migrate each resource with `terraform state rm` followed by `terraform import` instead. These commands change only the Terraform state and leave the networking object untouched.

In Terraform 1.12.0 and later, you can import with an `import` block and the `identity` attribute instead of the `terraform import` command. This way you do not need the `OS_*` and `INFRA_*` environment variables or the composite identifier: the `identity` attribute takes the project, the pool, and the ID of the object as separate fields. The fields of each resource are listed in the **Import** section of its page, for example, [`selectel_vpc_network_v2`](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_v2#import).

Before you start, [upgrade the version](#upgrade-the-version). Then migrate each resource with the procedure that matches your case:

- if the object is tracked in the current state — [Migrating a resource already in the current state](#migrating-a-resource-already-in-the-current-state)
- if the object was never tracked in this state, or is moving to another state or workspace — [Migrating never tracked resources](#migrating-never-tracked-resources)

#### Migrating a resource already in the current state

Take a network as an example. Your existing configuration looks like this:

```hcl
resource "openstack_networking_network_v2" "example" {
  name = "example"
}
```

1. Copy the network ID:

   ```bash
   terraform state show openstack_networking_network_v2.example
   ```

2. In the configuration files, replace the `openstack_networking_network_v2` resource with `selectel_vpc_network_v2` and add the `project_id` and `region` arguments. Keep the resource name unchanged. For example:

   ```hcl
   resource "selectel_vpc_network_v2" "example" {
     project_id = selectel_vpc_project_v2.project_1.id
     region     = "ru-3"
     name       = "example"
   }
   ```

3. If your OpenStack resource configuration sets any of the [attributes that become read-only or are not exposed](#attributes-that-become-read-only-or-are-not-exposed), remove them from the resource block. The Selectel resource does not accept them as arguments, so Terraform reports an unsupported-argument error before building a plan.

4. In the rest of the configuration, update references to the attributes that become read-only, and remove references to the attributes that are not exposed.

5. Remove the old resource address from the `.tfstate` file:

   ```bash
   terraform state rm openstack_networking_network_v2.example
   ```

   This does not delete the network.

6. Import the same network under the new resource type. Set the [environment variables](#environment-variables-for-import) first:

   ```bash
   export OS_DOMAIN_NAME=<account_id>
   export OS_USERNAME=<username>
   export OS_PASSWORD=<password>
   export INFRA_PROJECT_ID=<project_id>
   export INFRA_REGION=<region>
   terraform import selectel_vpc_network_v2.example <network_id>
   ```

   where `<network_id>` is the ID that you copied in step 1.

7. To ensure that Terraform does not plan to create or destroy the network, preview the changes:

   ```bash
   terraform plan
   ```

   For the attributes that the two schemas share, the plan shows no changes. Attributes that exist in only one of the two schemas are covered under [Schema differences](#schema-differences).

8. Apply the changes:

   ```bash
   terraform apply
   ```

9. Repeat steps 1–8 for each resource.

#### Migrating never tracked resources

Use `terraform import` when the object was never tracked in the `.tfstate` file or is moving to another state or workspace. The `terraform state rm` command is unnecessary.

1. In the configuration files, declare the `selectel_vpc_*` resource as it should look after the migration. For example:

   ```hcl
   resource "selectel_vpc_router_v2" "example" {
     project_id = selectel_vpc_project_v2.project_1.id
     region     = "ru-3"
     name       = "example"
   }
   ```

2. Copy the ID of the existing object. Any of these works:

   - if the resource is in a state file you can read — run `terraform state show openstack_networking_router_v2.example` and copy `id`
   - in the [Control panel](https://my.selectel.ru/), on the page of the object
   - with the OpenStack CLI against the same project, for example `openstack router list`
   - with a data source, where one exists — read `selectel_vpc_router_v2` by `name` and use its `router_id`

3. Import the object. Set the [environment variables](#environment-variables-for-import) first:

   ```bash
   export OS_DOMAIN_NAME=<account_id>
   export OS_USERNAME=<username>
   export OS_PASSWORD=<password>
   export INFRA_PROJECT_ID=<project_id>
   export INFRA_REGION=<region>
   terraform import selectel_vpc_router_v2.example <router_id>
   ```

   where `<router_id>` is the ID that you copied in step 2.

   A few resources use a composite import identifier instead of a plain one: `selectel_vpc_router_route_v2` and `selectel_vpc_network_subnet_route_v2` identify one entry of a list, so they take `<parent_id>-route-<destination_cidr>-<next_hop>`.

4. To ensure that the imported object matches the configuration, preview the changes:

   ```bash
   terraform plan
   ```

   For the attributes that the two schemas share, the plan shows no changes. Attributes that exist in only one of the two schemas are covered under [Schema differences](#schema-differences).

## Environment variables for import

The `terraform import` commands in this guide use these environment variables:

* `OS_DOMAIN_NAME` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [registration](https://docs.selectel.ru/en/account/registration/).
* `OS_USERNAME` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [service users](https://docs.selectel.ru/en/access-control/user-types/#service-users).
* `OS_PASSWORD` — Password of the service user.
* `INFRA_PROJECT_ID` — Unique identifier of the project. To get the ID, in the [Control panel](https://my.selectel.ru/), go to the product section in the navigation menu ⟶ project name ⟶ copy the ID of the required project. Learn more about [projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).
* `INFRA_REGION` — Pool, for example, `ru-6`. Learn more about available pools in [Product availability by location](https://docs.selectel.ru/en/infrastructure/product-availability-by-location/).

Selectel resources take `project_id` and `region` from `INFRA_PROJECT_ID` and `INFRA_REGION` during import rather than from the import string. The import fails if these variables are not set. `INFRA_REGION` is not needed for `selectel_vpc_public_subnet_v2`.
