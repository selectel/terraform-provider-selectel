---
layout: "selectel"
page_title: "Selectel: selectel_vpc_token_v2"
sidebar_current: "docs-selectel-resource-vpc-token-v2"
description: |-
  Creates and manages a Selectel Cloud Platform Keystone token using public API v2.
---

# selectel\_vpc\_token_v2

Creates and manages a Cloud Platform Keystone token using public API v2. For more information about Cloud Platform Keystone tokens, see the [official Selectel documentation](https://developers.selectel.ru/docs/control-panel/authorization/#токен-для-облачной-платформы-selectel).

> **WARNING**: This resource has been removed because it is for keystone tokens and they are automatically invalidated after 24 hours.

## Example Usage

```hcl
resource "selectel_vpc_token_v2" "token_1" {
  project_id = selectel_vpc_project_v2.project_1.id
}
```

## Argument Reference

* `project_id` - (Required) Unique identifier of the associated Cloud Platform project. Changing this creates a new Cloud Platform Keystone token. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

* `account_name` - (Optional) Selectel account ID. Changing this creates a new Cloud Platform Keystone token. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/).
