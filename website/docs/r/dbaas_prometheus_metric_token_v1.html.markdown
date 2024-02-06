---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_prometheus_metric_token_v1"
sidebar_current: "docs-selectel-resource-dbaas-prometheus-metric-token-v1"
description: |-
  Creates and manages tokens in Selectel Managed Databases required to get access to the metrics in the Prometheus format using public API v1.
---

# selectel\_dbaas\_prometheus_metric_token_v1

Creates and manages tokens required to get access to the metrics in the Prometheus format using public API v1. For more information about export of Prometheus metrics, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/cloud/managed-databases/postgresql/monitoring/#export-metrics-in-prometheus-format), [PostgreSQL for 1C](https://docs.selectel.ru/cloud/managed-databases/postgresql-for-1c/monitoring-1c/#export-metrics-in-prometheus-format), [PostgreSQL TimescaleDB](https://docs.selectel.ru/cloud/managed-databases/timescaledb/monitoring/#export-metrics-in-prometheus-format), [MySQL semi-sync](https://docs.selectel.ru/cloud/managed-databases/mysql-semi-sync/monitoring/#export-metrics-in-prometheus-format), [MySQL sync](https://docs.selectel.ru/cloud/managed-databases/mysql-sync/monitoring/#export-metrics-in-prometheus-format), [Redis](https://docs.selectel.ru/cloud/managed-databases/redis/monitoring/#export-metrics-in-prometheus-format) and [Kafka](https://docs.selectel.ru/cloud/managed-databases/kafka/monitoring/#export-metrics-in-prometheus-format).

## Example Usage

```hcl
resource "selectel_dbaas_prometheus_metric_token_v1" "token_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "token"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new token. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new token. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#облачные-базы-данных).

* `name` - (Required) Token name. Changing this creates a new token.

## Attributes Reference

* `value` (Sensitive) - Token value.

## Import

You can import a token:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export SEL_PROJECT_ID=<selectel_project_id>
export SEL_REGION=<selectel_pool>
terraform import selectel_dbaas_prometheus_metric_token_v1.token_1 <token_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<token_id>` — Unique identifier of the token, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the token ID in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ the cluster page ⟶ **Monitoring** tab ⟶ **Metrics in the Prometheus format** section ⟶ **Manage tokens**.