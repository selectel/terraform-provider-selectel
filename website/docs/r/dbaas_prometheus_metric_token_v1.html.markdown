---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_prometheus_metric_token_v1"
sidebar_current: "docs-selectel-resource-dbaas-prometheus-metric-token-v1"
description: |-
  Creates and manages tokens in Selectel Managed Databases required to get access to the metrics in the Prometheus format using public API v1.
---

# selectel\_dbaas\_prometheus_metric_token_v1

Creates and manages tokens required to get access to the metrics in the Prometheus format using public API v1. For more information about export of Prometheus metrics, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/cloud/managed-databases/postgresql/monitoring/#экспортировать-метрики-в-формате-prometheus), [PostgreSQL for 1C](https://docs.selectel.ru/cloud/managed-databases/postgresql-for-1c/monitoring-1c/#экспортировать-метрики-в-формате-prometheus), [PostgreSQL TimescaleDB](https://docs.selectel.ru/cloud/managed-databases/timescaledb/monitoring/#экспортировать-метрики-в-формате-prometheus), [MySQL semi-sync](https://docs.selectel.ru/cloud/managed-databases/mysql-semi-sync/monitoring/#экспортировать-метрики-в-формате-prometheus), [MySQL sync](https://docs.selectel.ru/cloud/managed-databases/mysql-sync/monitoring/#экспортировать-метрики-в-формате-prometheus), and [Redis](https://docs.selectel.ru/cloud/managed-databases/redis/monitoring/#экспортировать-метрики-в-формате-prometheus).

## Example Usage

```hcl
resource "selectel_dbaas_prometheus_metric_token_v1" "token_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
  name       = "token"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new token. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new token. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#облачные-базы-данных).

* `name` - (Required) Token name. Changing this creates a new token.

## Attributes Reference

* `value` (Sensitive) - Token value.

## Import

You can import a token:

```shell
$ export OS_DOMAIN_NAME=999999
$ export OS_USERNAME=example_user
$ export OS_PASSWORD=example_password
$ export SEL_PROJECT_ID=SELECTEL_VPC_PROJECT_ID
$ export SEL_REGION=SELECTEL_VPC_REGION
$ terraform import selectel_dbaas_prometheus_metric_token_v1.token <token_id>
```

where `<token_id>` is a unique identifier of the token, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the token ID in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ the cluster page ⟶ **Monitoring** tab ⟶ **Metrics in the Prometheus format** section ⟶ **Manage tokens**.

### Environment Variables

For import, you must set environment variables:

* `SEL_TOKEN=<selectel_api_token>`

* `SEL_PROJECT_ID=<selectel_project_id>`

* `SEL_REGION=<selectel_pool>`

where:

* `<selectel_api_token>` — Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶   **API keys**  ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel).

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to Cloud Platform ⟶ project name ⟶  copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-kubernetes/about/projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.
