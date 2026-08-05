---
layout: "selectel"
page_title: "Migrating networking resources from terraform-provider-openstack"
sidebar_current: "docs-selectel-guide-migrate-from-openstack-provider"
description: |-
  How to move existing Neutron networking resources from terraform-provider-openstack to the native networking resources of the Selectel provider.
---

# Migrating networking resources from terraform-provider-openstack

The Selectel provider now manages Neutron networking resources natively in its `vpc` domain. If you currently manage those resources through `terraform-provider-openstack`, this guide shows how to move them over without recreating anything in the cloud.

The Selectel cloud runs a patched Neutron with plugins that have no upstream equivalent, and it restricts who may change which attribute through granular VPC roles. The native resources are built for that; the upstream provider cannot be, because its schema assumes upstream Neutron and administrator privileges. That difference is what makes the migration worth doing — and it is also why a few attributes behave differently after you migrate. Those are listed under [Schema differences](#schema-differences).

~> **Note:** There is no automated migration tool. See [No automated tooling](#no-automated-tooling).

## Resource mapping

Each row maps one source resource type to the resource that replaces it. Every source type has exactly one counterpart.

| terraform-provider-openstack | terraform-provider-selectel | Data source |
| ---------------------------- | --------------------------- | ----------- |
| `openstack_networking_network_v2` | `selectel_vpc_network_v2` | `selectel_networking_network_v2` |
| `openstack_networking_subnet_v2` | `selectel_vpc_network_subnet_v2` | `selectel_networking_subnet_v2` |
| `openstack_networking_subnet_route_v2` | `selectel_vpc_network_subnet_route_v2` | none |
| `openstack_networking_port_v2` | `selectel_vpc_port_v2` | `selectel_networking_port_v2` |
| `openstack_networking_portforwarding_v2` | `selectel_vpc_port_forwarding_v2` | none |
| `openstack_networking_router_v2` | `selectel_vpc_router_v2` | `selectel_networking_router_v2` |
| `openstack_networking_router_interface_v2` | `selectel_vpc_router_interface_v2` | none |
| `openstack_networking_router_route_v2` | `selectel_vpc_router_route_v2` | none |
| `openstack_networking_router_routes_v2` | `selectel_vpc_router_routes_v2` | none |
| `openstack_networking_floatingip_v2` | `selectel_vpc_floatingip_v2` | `selectel_networking_floatingip_v2` |
| `openstack_networking_floatingip_associate_v2` | `selectel_vpc_floatingip_associate_v2` | none |
| `openstack_networking_secgroup_v2` | `selectel_vpc_secgroup_v2` | `selectel_networking_secgroup_v2` |
| `openstack_networking_secgroup_rule_v2` | `selectel_vpc_secgroup_rule_v2` | none |
| `openstack_networking_port_secgroup_associate_v2` | `selectel_vpc_port_secgroup_associate_v2` | none |
| `openstack_networking_rbac_policy_v2` | `selectel_vpc_rbac_policy_v2` | none |

Six of the fifteen entities get a data source, listed in the third column. The others are primarily link resources — a route entry, an interface, an association, or a policy — with no independent object to search for.

~> **Note:** Floating IP allocation remains managed by `selectel_vpc_floatingip_v2`. The read-only `selectel_networking_floatingip_v2` data source exposes its Neutron state, while `selectel_vpc_floatingip_associate_v2` manages the association with a port.

### Resources this guide does not cover

The following are **not** part of this migration and stay under `terraform-provider-openstack`. You will not find rows for them above:

* Neutron trunks and network segments
* Quotas
* QoS policies and QoS rules (bandwidth-limit, DSCP-marking, minimum-bandwidth)
* BGP peers and BGP speakers
* FWaaS resources — firewall groups, policies and rules
* `bmnet` devices, and the `bmnet:*` attributes on a network
* **Address groups**, and the `remote_address_group_id` attribute of a security group rule

Address groups deserve a word, because the upstream provider does offer `openstack_networking_address_group_v2`. In the Selectel cloud, creating, updating and deleting an address group is restricted to the cloud administrator: the deployed Neutron policy registers only `get_address_group`, and an unregistered policy rule denies the action rather than allowing it. Reading existing address groups is permitted, so a future change may add a data source, but there is nothing here for a project member to manage. Because a security group rule can only usefully reference an address group you can create, `remote_address_group_id` is left out with it.

~> **Note:** Nothing in your existing configuration is deleted by this migration. Resources that stay with the upstream provider keep working alongside the native ones; both talk to the same Neutron.

## Migration procedure

Changing from `openstack_networking_*` to `selectel_vpc_*` changes the Terraform resource type. `terraform state mv` cannot do that: its source and destination must have the same resource type. Use `terraform state rm` followed by `terraform import` instead. These commands change only Terraform state and leave the Neutron object untouched.

## Migrating an object already in the current state

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

This does not delete the network in Neutron. It only removes its old provider binding from state.

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

## Migrating with `terraform import`

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
* With a data source, where one exists — read `selectel_networking_router_v2` by `name` and use its `router_id`.

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

## Schema differences

The native resources are not a field-for-field copy of the upstream ones, in two directions. Import does not fill these values into your configuration, so importing the object is not by itself a complete transfer of functionality.

### Attributes that are new, and read-only

The Selectel Neutron carries attributes the upstream schema has no place for. They report platform state; the provider reflects them and never sends them, so there is nothing for you to add to your configuration:

| Entity | New read-only attributes |
| ------ | ------------------------ |
| network | `blocked`, `is_public`, `is_dns_enabled` |
| subnet | `blocked` |
| port | `blocked`, `dhcp_blocked` |
| router | `blocked` |

A security group has no `blocked` attribute. It does gain a read-only `shared`, which reports whether an RBAC policy grants another project access to the group; you manage that with `selectel_vpc_rbac_policy_v2`, not on the group itself.

After a migration these appear in your state with the values the platform reports. A change made to one of them outside Terraform — an administrator blocking a network, say — is recorded on the next refresh and does not produce a permanent diff.

### Attributes the upstream provider lets you set, and these do not

The cloud's Neutron policy reserves some attributes for the Keystone admin role, granting them to none of the granular VPC roles. Where the upstream schema accepts them as arguments, the native schema exposes them read-only, or not at all.

~> **Warning:** Remove these lines from your configuration as you migrate. Left in place, `terraform plan` fails with an unsupported-argument error, because the attribute is not a configuration argument of the new resource. The actual values stay under the platform's control.

| Entity | Read-only in the native resource | Not exposed at all |
| ------ | -------------------------------- | ------------------ |
| network | `shared`, `external`, `mtu`, `port_security_enabled`, `qos_policy_id`, `provider:network_type`, `provider:physical_network`, `provider:segmentation_id` | `segments` (replaced by the three read-only `provider_*` attributes) |
| subnet | — | `service_types`, `segment_id` |
| port | `port_security_enabled`, `qos_policy_id` | `binding.host_id`, `binding.profile` |
| router | `external_fixed_ip`, gateway `qos_policy_id` | `distributed`, `ha`, `external_subnet_ids` |
| subnet pool | `shared`, `is_default` | — |
| address scope | `shared` | — |

Two of these are worth expanding on:

* **Router `distributed` and `ha`** are not exposed even as read-only values. The policy reserves not only writing them but *reading* them for the administrator, so an ordinary caller never receives the fields at all — reporting `false` would state something the provider has not been told.

### Behaviour that differs from upstream Neutron

Two behaviours come from patches in the Selectel Neutron and will differ from what you are used to, whichever migration path you take:

* **Router availability zones.** A caller without the administrator role may only use `availability_zone_hints` values from the region's allowlist. A value outside a non-empty allowlist fails with an authorization error; an unset allowlist fails with a not-found error.
* **Static route next hops.** A router's ports on external networks are excluded when Neutron decides which next hops are acceptable. An address inside the router's external gateway subnet is therefore refused, even though upstream Neutron accepts it. If you have such a route today under the upstream provider, it will not apply here.

## No automated tooling

This change ships no migration tool: no generator for `state rm`/`import` commands, no bulk-import script, no conversion utility for `.tf` files.

Migration is manual and per-resource. If you manage dozens of subnets, you run the `state rm` and `import` steps for each of them. That is the intended workflow; check every imported object with `terraform plan` before applying anything.
