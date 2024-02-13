---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_kafka_ccl_v1"
sidebar_current: "docs-selectel-resource-dbaas-kafka-acl-v1"
description: |-
  Creates and manages an ACL in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_kafka\_acl\_v1

Creates and manages an access control list (ACL) in a Kafka datastore using public API v1. For more information about managing users in Kafka, see the [official Selectel documentation](https://docs.selectel.ru/cloud/managed-databases/kafka/manage-users/)

## Example usage

```hcl
resource "selectel_dbaas_kafka_acl_v1" "acl_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_kafka_datastore_v1.datastore_1.id
  pattern      = "topic"
  pattern_type = "prefixed"
  allow_read   = true
  allow_write  = true
}
```

## Argument Reference

* `pattern` - (Optional) Name or prefix of a topic to which you provide access. Changing this creates a new ACL. Must be skipped when `pattern_type` is `all`.

* `pattern_type` - (Required) ACL pattern type. Changing this creates a new ACL. Available ACL patterns are `prefixed`, `literal`, and  `all`. When `pattern_type` is `all`, skip pattern.

* `allow_read` - (Required) Allows to connect as a consumer.

* `allow_write` - (Required) Allows to connect as a producer.

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new user. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new ACL. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#managed-databases).

* `datastore_id` - (Required) Unique identifier of the associated datastore. Changing this creates a new ACL. Retrieved from the [selectel_dbaas_kafka_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_kafka_datastore_v1).

* `user_id` - (Required) Unique identifier of the associated user. Changing this creates a new ACL. Retrieved from the [selectel_dbaas_user_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_user_v1) resource.

## Attributes Reference

* `status` - ACL status.
