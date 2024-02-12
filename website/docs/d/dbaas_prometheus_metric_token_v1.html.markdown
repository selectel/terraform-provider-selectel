---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_prometheus_metric_token_v1"
sidebar_current: "docs-selectel-datasource-dbaas-prometheus-metric-token-v1"
description: |-
  Provides a list of tokens for Prometheus available in Selectel Managed Databases.
---

# selectel\_dbaas\_prometheus_metric_token_v1

Provides a list of tokens for Prometheus available in Managed Databases. For more information about tokens for Prometheus, see the official Selectel documentation for [PostgreSQL](https://docs.selectel.ru/cloud/managed-databases/postgresql/monitoring/#export-metrics-in-prometheus-format), [PostgreSQL for 1C](https://docs.selectel.ru/cloud/managed-databases/postgresql-for-1c/monitoring-1c/#export-metrics-in-prometheus-format), [PostgreSQL TimescaleDB](https://docs.selectel.ru/cloud/managed-databases/timescaledb/monitoring/#export-metrics-in-prometheus-format), [MySQL semi-sync](https://docs.selectel.ru/cloud/managed-databases/mysql-semi-sync/monitoring/#export-metrics-in-prometheus-format), [MySQL sync](https://docs.selectel.ru/cloud/managed-databases/mysql-sync/monitoring/#export-metrics-in-prometheus-format), [Redis](https://docs.selectel.ru/cloud/managed-databases/redis/monitoring/#export-metrics-in-prometheus-format), and [Kafka](https://docs.selectel.ru/cloud/managed-databases/kafka/monitoring/#export-metrics-in-prometheus-format).

## Example Usage

```hcl
data "selectel_dbaas_prometheus_metric_token_v1" "token_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-3"
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#облачные-базы-данных).

## Attributes Reference

* `prometheus_metrics_tokens` -  List of tokens for Prometheus.

  * `id` - Unique identifier of the token.

  * `created_at` - Time when the token was created.

  * `updated_at` - Time when the token was updated.

  * `project_id` - Unique identifier of the associated Cloud Platform project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

  * `name` - Token name.

  * `value` - Token value.