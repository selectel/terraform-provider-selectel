---
layout: "selectel"
page_title: "Selectel: selectel_secretsmanager_secret_v1"
sidebar_current: "docs-selectel-resource-secretsmanager-secret-v1"
description: |-
    Creates and manages a secret in Selectel Secrets Manager using public API v1.
---

# selectel\_secretsmanager\_secret_v1

Creates and manages a secret in Selectel Secrets Manager using public API v1. For more information about Secrets Manager, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/secrets-manager/secrets/).

## Example Usage

```hcl
resource "selectel_secretsmanager_secret_v1" "secret_1" {
  key         = "secret"
  value       = "verysecret"
  project_id  = selectel_vpc_project_v2.project_1.id
  description = "secret from .tf"
}
```

## Argument Reference

* `key` - (Required) Secret name.

* `value` - (Required, Sensitive) Secret value, for example password, API key, certificate key. The limit is 65 536 characters.

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `description` - (Optional) Secret description.

## Attributes Reference

* `created_at` - Time when the secret was created.

* `name` - Secret name, same as the secret key.

## Import

You can import a secret:

```shell
export INFRA_PROJECT_ID=<selectel_project_id>
terraform import selectel_secretsmanager_secret_v1.secret_1 <selectel_project_id>/<key>
```

where:

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/secrets-manager), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<key>` — Secret name. To get the secret name, in the [Control panel](https://my.selectel.ru/vpc/secrets-manager/), go to **Cloud Platform** ⟶ **Secrets Manager** ⟶ the **Secrets** tab ⟶ copy the name of the required secret.