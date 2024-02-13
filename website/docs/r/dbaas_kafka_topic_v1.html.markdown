---
layout: "selectel"
page_title: "Selectel: selectel_dbaas_kafka_topic_v1"
sidebar_current: "docs-selectel-resource-dbaas-kafka-topic-v1"
description: |-
  Creates and manages a topic in Selectel Managed Databases using public API v1.
---

# selectel\_dbaas\_kafka\_topic\_v1

Creates and manages a topic in a Kafka datastore using public API v1. For more information about managing topics in Kafka, see the [official Selectel documentation](https://docs.selectel.ru/cloud/managed-databases/kafka/manage-topics/)

## Example usage

```hcl
resource "selectel_dbaas_kafka_topic_v1" "topic_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-3"
  datastore_id = selectel_dbaas_kafka_datastore_v1.datastore_1.id
  name         = "topic"
  partitions   = 1
}
```

## Argument Reference

* `name` - (Required, Sensitive) Topic name. Changing this creates a new topic.

* `partitions` - (Required) Number of partitions in a topic. The available range is from 1 to 4 000. You cannot increase the number of partitions in the existing topic. Learn more about [Partitions](https://docs.selectel.ru/cloud/managed-databases/kafka/manage-topics/#partitions)

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new topic. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `region` - (Required) Pool where the database is located, for example, `ru-3`. Changing this creates a new topic. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/#managed-databases).

* `datastore_id` - (Required) Unique identifier of the associated datastore. Changing this creates a new topic. Retrieved from the [selectel_dbaas_kafka_datastore_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/dbaas_kafka_datastore_v1).

## Attributes Reference

* `status` - Topic status.

## Import

You can import a topic:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export SEL_PROJECT_ID=<selectel_project_id>
export SEL_REGION=<selectel_pool>
terraform import selectel_dbaas_kafka_topic_v1.topic_1 <topic_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `<selectel_pool>` — Pool where the cluster is located, for example, `ru-3`. To get information about the pool, in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases**. The pool is in the **Pool** column.

* `<topic_id>` — Unique identifier of the topic, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. To get the topic ID in the [Control panel](https://my.selectel.ru/vpc/dbaas/), go to **Cloud Platform** ⟶ **Managed Databases** ⟶ the cluster page ⟶ the **Topics** tab. The topic ID is under the topic name.