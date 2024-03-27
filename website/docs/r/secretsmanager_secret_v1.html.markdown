---
layout: "selectel"
page_title: "Selectel: selectel_secretsmanager_secret_v1"
sidebar_current: "docs-selectel-resource-secretsmanager-secret-v1"
description: |-
    Creates and manages a Secret in Selectel SecretsManager service using public API v1.
---

# selectel\_secretsmanager\_secret_v1

Creates and manages a Secret in Selectel SecretsManager service using public API v1. Learn more about [Secrets](https://docs.selectel.ru/en/cloud/secrets-manager/secrets/).

## Example Usage
```hcl
resource "selectel_secretsmanager_secret_v1" "secret_1" {
    key = "Terraform-Secret"
    description = "Secret from .tf"
    value = "zelibobs"
    project_id = selectel_vpc_project_v2.project_1.id
}
```

## Argument Reference
- `key` (Required) — unique key, name of the secret.
- `description` (Optinal) — description of the secret.
- `value` (Required, Sensitive) — secret value, e.g. password, API key, certificate key, or other.
- `project_id` (Required) — unique identifier of the associated Cloud Platform project.

## Attributes Reference
- `created_at` — time when the secret was created.
- `name` — computed name of the secret same as key.

## Import

~> When importing Secret you have to provide unique identifier of the associated Cloud Platform project

### Using import block
-> In Terraform v1.5.0 and later, use an import block to import Secret using template below.

```hcl
import {
   to = selectel_secretsmanager_secret_v1.imported_secret
   id = "<selectel_project_id>/<key>"
}
```

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `<key>` — Unique identifier of the secret, its name. To get the name of the secret in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Secrets Manager** ⟶ **Secret** copy the Name.



### Using terraform import
```shell
export SEL_PROJECT_ID=<selectel_project_id>
terraform import selectel_secretsmanager_secret_v1.imported_secret <selectel_project_id>/<key>
```

* `<selectel_project_id>` — Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/managed-databases/about/projects/).

* `<key>` — Unique identifier of the secret, its name. To get the name of the secret in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Secrets Manager** ⟶ **Secret** copy the Name.
