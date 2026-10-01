---
page_title: "Upgrading to version 9.0.0"
description: |-
  How to upgrade Selectel Terraform Provider to version 9.0.0, migrate the deprecated public subnet resource, and move existing networking resources from terraform-provider-openstack.
---

# Upgrading to version 9.0.0

Version 9.0.0 of the Selectel provider introduces `selectel_vpc_public_subnet_v2` as a replacement for the deprecated `selectel_vpc_subnet_v2` resource and manages networking resources natively in its `vpc` (Virtual Private Cloud) domain. This guide explains how to migrate both kinds of resources without recreating them.

The VPC API behavior can differ from the OpenStack API. The differences relevant to migration are listed under [Schema differences](#schema-differences).

## Migrating the deprecated public subnet resource

The `selectel_vpc_subnet_v2` resource is deprecated. Existing configurations continue to work, but use `selectel_vpc_public_subnet_v2` for new configurations. Both resource types manage the same kind of public subnet and use the same arguments and attributes.

Terraform cannot use [`terraform state mv`](https://developer.hashicorp.com/terraform/cli/commands/state/mv) when the source and destination have different resource types. Migrate the state with `terraform state rm` and `terraform import` instead. These commands change only Terraform state and do not recreate the public subnet.

Before you start, back up the state and record the public subnet ID:

```bash
terraform state show selectel_vpc_subnet_v2.example
```

Change the resource type in the configuration and update every reference to it. Keep the resource name and arguments unchanged:

```hcl
resource "selectel_vpc_public_subnet_v2" "example" {
  project_id    = selectel_vpc_project_v2.project_1.id
  region        = "ru-3"
  ip_version    = "ipv4"
  prefix_length = 29
}
```

Remove the old resource address from the state:

```bash
terraform state rm selectel_vpc_subnet_v2.example
```

Import the same public subnet under the new resource type:

```bash
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
terraform import selectel_vpc_public_subnet_v2.example <public_subnet_id>
```

Run `terraform plan` and make sure Terraform does not plan to create or destroy the public subnet.

~> **Warning:** Do not declare both resource types for the same public subnet. Do not run `terraform apply` between `terraform state rm` and `terraform import`.

## Migrating resources from terraform-provider-openstack

### Resource mapping

Each row maps one source resource type to the resource that replaces it. Every source type has exactly one counterpart.

| terraform-provider-openstack | terraform-provider-selectel | Data source |
| ---------------------------- | --------------------------- | ----------- |
| `openstack_networking_network_v2` | `selectel_vpc_network_v2` | `selectel_vpc_network_v2` |
| `openstack_networking_subnet_v2` | `selectel_vpc_network_subnet_v2` | `selectel_vpc_network_subnet_v2` |
| `openstack_networking_subnet_route_v2` | `selectel_vpc_network_subnet_route_v2` | none |
| `openstack_networking_port_v2` | `selectel_vpc_port_v2` | `selectel_vpc_port_v2` |
| `openstack_networking_portforwarding_v2` | `selectel_vpc_port_forwarding_v2` | none |
| `openstack_networking_router_v2` | `selectel_vpc_router_v2` | `selectel_vpc_router_v2` |
| `openstack_networking_router_interface_v2` | `selectel_vpc_router_interface_v2` | none |
| `openstack_networking_router_route_v2` | `selectel_vpc_router_route_v2` | none |
| `openstack_networking_router_routes_v2` | `selectel_vpc_router_routes_v2` | none |
| `openstack_networking_floatingip_associate_v2` | `selectel_vpc_floatingip_associate_v2` | none |
| `openstack_networking_secgroup_v2` | `selectel_vpc_secgroup_v2` | `selectel_vpc_secgroup_v2` |
| `openstack_networking_secgroup_rule_v2` | `selectel_vpc_secgroup_rule_v2` | none |
| `openstack_networking_port_secgroup_associate_v2` | `selectel_vpc_port_secgroup_associate_v2` | none |
| `openstack_networking_rbac_policy_v2` | `selectel_vpc_rbac_policy_v2` | none |

Five of the fourteen entities get a data source, listed in the third column. The others either represent links without an independent object to search for or, in the case of floating IPs, keep allocation and deletion in the existing resource.

Networking resources that are not listed in the [Resource mapping](#resource-mapping) are not supported in the Selectel VPC.

### Migrating data sources

Data sources do not manage the lifecycle of remote objects, so they do not need `terraform state rm` or `terraform import`. Replace the OpenStack data source type with the corresponding Selectel data source from the mapping table, add the required `project_id` and `region` selectors, and update references to its attributes.

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

Run `terraform plan` to verify that the new data source selects the expected object and that all references use attributes available in its schema.

### Migrating managed resources

Changing from `openstack_networking_*` to `selectel_vpc_*` changes the Terraform resource type. `terraform state mv` cannot do that: its source and destination must have the same resource type. Use `terraform state rm` followed by `terraform import` instead. These commands change only Terraform state and leave the networking resource untouched in the VPC.

### Migrating an object already in the current state

Before you start, make sure the `selectel` provider is configured in your configuration, and back up your state file.

Take a network as the example. Your existing configuration and state look like this:

```hcl
resource "openstack_networking_network_v2" "example" {
  name = "example"
}
```

**Step 1 — update the configuration.** Change the resource type, and add the `project_id` and `region` the native resources require:

```hcl
resource "selectel_vpc_network_v2" "example" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "example"
}
```

**Step 2 — stop tracking the old resource address:**

```bash
terraform state rm openstack_networking_network_v2.example
```

This does not delete the network in the VPC. It only removes its old provider binding from state.

**Step 3 — import the same object under the new type:**

```bash
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_vpc_network_v2.example <network_id>
```

Record the object ID with `terraform state show` before removing the old binding.

**Step 4 — check the plan:**

```bash
terraform plan
```

For the attributes the two schemas share, the plan shows no destroy and no create: the same object is now managed under the new type, with the same id. Attributes that exist in only one of the two schemas are covered under [Schema differences](#schema-differences).

Repeat for each resource, then run `terraform apply` once you are satisfied with the plan.

~> **Warning:** Back up the state first, record the remote object ID, and do not run `terraform apply` between `state rm` and `import`. Updating the `.tf` file is mandatory: leaving the old resource block would make Terraform plan a new object under the old type.

### Migrating with `terraform import`

Use the same import operation when the object was never tracked in this state or is moving to another state/workspace; the `state rm` step is simply unnecessary.

**Step 1 — write the configuration** for the native resource, as it should look after the migration:

```hcl
resource "selectel_vpc_router_v2" "example" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "example"
}
```

**Step 2 — find the identifier of the existing object.** Any of these works:

* If the resource is in a state file you can read: `terraform state show openstack_networking_router_v2.example` and take `id`.
* From the [Control panel](https://my.selectel.ru/), on the page of the object.
* With the OpenStack CLI against the same project — for example `openstack router list`.
* With a data source, where one exists — read `selectel_vpc_router_v2` by `name` and use its `router_id`.

**Step 3 — import it:**

```bash
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>
terraform import selectel_vpc_router_v2.example 1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809
```

The native resources take `project_id` and `region` from `INFRA_PROJECT_ID` and `INFRA_REGION` during import rather than from the import string, and the import fails if those variables are not set. A few resources use a composite import identifier instead of a plain one — `selectel_vpc_router_route_v2` and `selectel_vpc_network_subnet_route_v2` identify one entry of a list, so they take `<parent_id>-route-<destination_cidr>-<next_hop>`. Each resource page documents its own import format.

**Step 4 — check the plan.** With a matching configuration, `terraform plan` shows no changes for the attributes the two schemas share.

### Schema differences

#### Attributes that are new, and read-only

The Selectel VPC carries attributes the upstream schema has no place for. They report platform state; the provider reflects them and never sends them, so there is nothing for you to add to your configuration:

| Entity | New read-only attributes |
| ------ | ------------------------ |
| network | `status`, `subnets`, `availability_zones`, `blocked`, `is_public`, `is_dns_enabled` |
| subnet | `host_routes`, `blocked` |
| port | `status`, `blocked`, `dhcp_blocked` |
| router | `status`, `availability_zones`, `blocked` |
| security group | `shared`, `security_group_rules` |

A security group has no `blocked` attribute. Its `shared` attribute reports whether an RBAC policy grants another project access to the group; manage that access with `selectel_vpc_rbac_policy_v2`, not on the group itself. `security_group_rules` contains the identifiers of the rules that belong to the group.

#### Attributes that become read-only or are not exposed

~> **Warning:** If your OpenStack resource configuration sets any attributes listed below, remove them when changing the resource type. The Selectel resource does not accept them as configuration arguments, so Terraform reports an unsupported-argument error before building a plan. Read-only attributes remain available for references; attributes listed as not exposed must also be removed from references.

| Entity | Read-only in the native resource | Not exposed at all |
| ------ | -------------------------------- | ------------------ |
| network | `shared`, `external`, `mtu`, `port_security_enabled` | `tenant_id`, `segments`, `transparent_vlan`, `value_specs`, `qos_policy_id` |
| subnet | `service_types` | `tenant_id`, `prefix_length`, `dns_publish_fixed_ip`, `segment_id`, `subnetpool_id`, `value_specs` |
| port | `port_security_enabled` | `tenant_id`, `value_specs`, `binding`, `dns_assignment`, `qos_policy_id` |
| router | `external_fixed_ip` | `tenant_id`, `external_qos_policy_id`, `distributed`, `external_subnet_ids`, `flavor_id`, `value_specs`, `vendor_options` |
| security group | — | `tenant_id` |

The native resources use the required `project_id` argument instead of `tenant_id`. Remove `tenant_id` rather than copying it alongside `project_id`.

#### Other schema changes

* Network `availability_zone_hints` changes from an unordered `Set` to an ordered `List`.
* Subnet `service_types` changes from a configurable `List` to a read-only `Set`. Remove it from the managed subnet configuration, and do not use numeric indexes when referencing it.
* Subnet `prefix_length` and subnet-pool allocation are not supported. The OpenStack resource uses `prefix_length` with `subnetpool_id` to allocate a CIDR automatically; the native resource requires an explicit `cidr`. Before migration, copy the allocated CIDR from the existing resource state into the new configuration.
* Port `all_fixed_ips` changes from a list of IP address strings to a list of objects with `subnet_id` and `ip_address`. For example, replace `all_fixed_ips[0]` with `all_fixed_ips[0].ip_address`.
* Security group `description` is no longer computed when omitted. Copy a non-empty existing description into the configuration if it must be preserved.

#### Changed defaults

`admin_state_up` defaults to `true` for managed networks, ports, and routers. To preserve an imported object that is administratively down, set `admin_state_up = false` explicitly; otherwise Terraform plans to enable it.

### No automated tooling

This change ships no migration tool: no generator for `state rm`/`import` commands, no bulk-import script, no conversion utility for `.tf` files.

Migration is manual and per-resource. If you manage dozens of subnets, you run the `state rm` and `import` steps for each of them. That is the intended workflow; check every imported object with `terraform plan` before applying anything.
