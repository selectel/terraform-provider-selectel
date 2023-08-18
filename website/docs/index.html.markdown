---
layout: "selectel"
page_title: "Provider: Selectel"
sidebar_current: "docs-selectel-index"
description: |-
  Use the Selectel provider to interact with Selectel products.
---

# Selectel provider

Use the Selectel Terraform provider to interact with [Selectel products](https://docs.selectel.ru/).

To interact with resources that are available via OpenStack API, you can also use [OpenStack Terraform provider](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest).

## Example Usage

```hcl
terraform {
  required_providers {
    selectel = {
      source = "selectel/selectel"
      version = "~> 3.11.0"
    }
  }
}

# Create a Cloud Platform project
resource "selectel_vpc_project_v2" "project_1" {
  ...
}
```

## Authentication

```hcl
# Configure the Selectel Provider

provider "selectel" {
  token = "<selectel_token>"
}
```

## Argument Reference

* `token` - (Required) Selectel token. To get the token, in the top right corner of the [Control panel](https://my.selectel.ru/profile/apikeys), go to the account menu ⟶ **Profile and Settings** ⟶ **API keys** ⟶ copy the token. Learn more about [Selectel token](https://developers.selectel.ru/docs/control-panel/authorization/#получить-токен-selectel). If skipped, use the `SEL_TOKEN` environment variable.

* `endpoint` - (Optional) Selectel API endpoint. Use only for test environments. If skipped, the provider automatically uses the official Selectel endpoint.

* `project_id` - (Optional) Unique identifier of the Cloud Platform project. Use only to import resources that are associated with the specific project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/), go to the **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. As an alternative, you can retrieve project ID from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/). If skipped, use the `SEL_PROJECT_ID` environment variable. 

* `region` - (Optional) Pool, for example, `ru-3`. Use only to import resources from the specific pool. Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/control-panel-actions/availability-matrix/). If skipped, use the `SEL_REGION` environment variable.
