---
page_title: "Selectel: selectel_secretsmanager_certificate_v1"
description: |-
    Creates and manages a certificate in Selectel Secrets Manager using public API v1.
---

# selectel\_secretsmanager\_certificate_v1

Creates and manages a certificate in Selectel Secrets Manager using public API v1. For more information about certificates, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/secrets-manager/certificates/).

## Example Usage

### Text format

```terraform
resource "selectel_secretsmanager_certificate_v1" "certificate_1" {
  name          = "certificate"
  certificates  = [file("./_cert.pem")]
  private_key   = file("./_private_key.pem")
  project_id    = selectel_vpc_project_v2.project_1.id
}
```

### EOF

```terraform
resource "selectel_secretsmanager_certificate_v1" "certificate_1" {
  name         = "certificate"
  certificates = [
      <<-EOF
      -----BEGIN CERTIFICATE-----
      ...
      ----END CERTIFICATE-----
      EOF
  ]
  private_key  = <<-EOF
  -----BEGIN PRIVATE KEY-----
  ...
  -----END PRIVATE KEY-----
  EOF
  project_id   = selectel_vpc_project_v2.project_1.id
}
```

## Argument Reference

* `name` - (Required) Certificate name.

* `certificates` - (Required) Certificate chain in PEM format. The value of each certificate must begin with `-----BEGIN CERTIFICATE-----` and end with `-----END CERTIFICATE-----`.

* `private_key` - (Required, Sensitive) Private key. The value must begin with `-----BEGIN PRIVATE KEY-----` and end with `-----END PRIVATE KEY-----`.

* `project_id` - (Required) Unique identifier of the associated project. Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `current_version` - (Optional) Desired active version. Defaults to `0` (automatic activation on certificate content changes). Positive values select an existing version. Must be non-negative; only `0` or `1` is valid when creating a certificate.

## Version Management

Changing `certificates` or `private_key` creates a new version. With `current_version = 0`, the new version becomes active. Changing only `name` does not create a version.

To roll back, set `current_version` to an existing version number and apply. Keep `certificates` and `private_key` unchanged during rollback: they represent the last content uploaded by Terraform, not necessarily the active version's content.

When `current_version` is positive, changing certificate content uploads an **inactive** version and keeps the selected version active. To activate the uploaded version, set `current_version` to its version number in a subsequent apply. Setting `current_version` back to `0` leaves the active version unchanged; subsequent content changes activate their new versions.

`version` and the computed certificate metadata report the observed active version, not the latest uploaded version. With a positive `current_version`, refresh detects external activation changes and the next apply restores the configured version. With `0`, external activation changes are accepted.

Versions cannot be deleted individually through this resource. Destroying the resource deletes the entire certificate, including its versions.

### Pin or roll back an existing certificate

```terraform
# Apply this change to an existing certificate. Keep its last uploaded PEM unchanged.
# Version 1 also exists immediately after creation.
resource "selectel_secretsmanager_certificate_v1" "certificate_1" {
  name            = "certificate"
  certificates    = [file("./_cert.pem")]
  private_key     = file("./_private_key.pem")
  project_id      = selectel_vpc_project_v2.project_1.id
  current_version = 1
}

# Changing PEM while pinned uploads an inactive version.
# Set current_version to that version number in a subsequent apply to activate it.
# Set current_version to 0 to automatically activate future content changes.
```

## Attributes Reference

* `dns_names` - Domain names for which the certificate is issued.

* `id` - Unique identifier of the certificate.

* `issued_by` - Information about the Certificate Authority (CA) which verified and signed the certificate.

* `serial` - Certificate serial number assigned by the Certificate Authority (CA) which issued the certificate.

* `validity` - Certificate validity in the RFC3339 timestamp format:

  * `not_before` - Effective date and time of the certificate.

  * `not_after` - Expiration date and time of the certificate.

* `version` - Active certificate version number (computed).

## Import

You can import a certificate:

```shell
export INFRA_PROJECT_ID=<selectel_project_id>
terraform import selectel_secretsmanager_certificate_v1.certificate_1 <certificate_id>
```

where:

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/secrets-manager), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<certificate_id>` — Unique identifier of the certificate. To get the ID of the certificate, in the [Control panel](https://my.selectel.ru/vpc/secrets-manager/), go to **Cloud Platform** ⟶ **Secrets Manager** ⟶ the **Certificates** tab ⟶ in the certificate menu select **Copy UUID**.
