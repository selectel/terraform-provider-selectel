---
page_title: "Selectel: selectel_secretsmanager_secret_v1"
description: |-
  Creates and manages a secret in Secrets Manager using public API v1.
---

# selectel\_secretsmanager\_secret\_v1

Creates and manages a secret in Secrets Manager using public API v1. For more information about Secrets Manager, see the [official Selectel documentation](https://docs.selectel.ru/en/secrets-manager/secrets/).

## Example Usage

### Create a secret

Create `./secret.txt` with a non-empty UTF-8 secret value before applying.

```terraform
resource "selectel_secretsmanager_secret_v1" "secret_1" {
  key        = "secret"
  value      = file("./secret.txt")
  project_id = selectel_vpc_project_v2.project_1.id
}
```

### Upload a new version without activating it

```terraform
# Update the value of an existing secret to upload a version without activating it.
resource "selectel_secretsmanager_secret_v1" "secret_1" {
  key        = "secret"
  value      = file("./secret.txt")
  project_id = selectel_vpc_project_v2.project_1.id
  is_current = false
}
```

## Argument Reference

* `key` - (Required) Secret name, unique within the project. Must contain 1 to 255 ASCII letters, digits, underscores, or hyphens (`[a-zA-Z0-9_-]`). Changing this argument recreates the secret and deletes all its versions.

* `value` - (Required, Sensitive) Non-empty secret value as plain UTF-8 text, for example a password, API key, or certificate key. Maximum size is 262,144 bytes (256 KiB). Supply plain text or use `file("./secret.txt")`; the provider handles API encoding. Changing this argument creates a new version, whose activation depends on `is_current`. Sensitive values are hidden in normal Terraform output but remain stored in Terraform state.

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/). Changing this creates a new secret and deletes the existing secret and all its versions.

* `description` - (Optional) Secret description. Defaults to an empty string. Maximum length is 1024 characters. Each character must use at most three UTF-8 bytes; four-byte characters, such as most emoji, are not supported.

* `is_current` - (Optional) Whether to activate a new version when you update the resource content. The first version is always active. Changing only this argument does not switch the active version. The default value is `true`.

## Version Management

To activate an existing version, use Secrets Manager outside Terraform. The `value` argument retains the last content uploaded by Terraform. Learn more about [Secret versioning](https://docs.selectel.ru/en/secrets-manager/secrets/versioning/).

Versions cannot be deleted individually through this resource. Destroying the resource deletes the entire secret, including its versions.

## Attributes Reference

* `id` - Unique identifier of the secret in the `<project_id>/<key>` format.

* `created_at` - Time when the active secret version was created in RFC3339 format.

* `name` - Secret name, same as the secret key.

* `version` - Number of the active version, which may differ from the latest uploaded version.

## Import

Import does not retrieve the existing secret's plaintext value. If value is missing from state, the provider sets it to the literal string `UNKNOWN`. Set the required value in the resource configuration before applying. If the configured value differs from the imported state, apply creates a new secret version; whether that version is activated depends on `is_current`. Import does not recover the original value and does not guarantee that the first apply will leave the secret unchanged.

You can import a secret:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export OS_AUTH_URL=<auth_url>
export OS_REGION_NAME=<auth_region>
export INFRA_PROJECT_ID=<selectel_project_id>
terraform import selectel_secretsmanager_secret_v1.secret_1 <selectel_project_id>/<key>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/service-users), go to **IAM** ⟶ **Service users** ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/en/access-control/users/about/user-types/).

* `<password>` — Password of the service user.

* `<auth_url>` — Authentication endpoint used by the provider. Obtain it from the [list of URLs](https://docs.selectel.ru/en/api/urls/).

* `<auth_region>` — Authentication region matching the selected authentication endpoint. Use the value from your configured provider.

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/iam/projects), go to **IAM** ⟶ **Projects** ⟶ copy the ID of the required project. The configured project ID must match the project ID in the import identifier. The service user must have access to this project. You can set the provider's `project_id` argument instead of `INFRA_PROJECT_ID`. Learn more about [Projects](https://docs.selectel.ru/en/access-control/projects/about-projects/).

* `<key>` — Secret name. In the [Control panel](https://my.selectel.ru/secrets/), open **Products** ⟶ **Secrets Manager** ⟶ **Secrets** and copy the existing secret's name.
