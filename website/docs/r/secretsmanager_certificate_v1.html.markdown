---
layout: "selectel"
page_title: "Selectel: selectel_secretsmanager_certificate_v1"
sidebar_current: "docs-selectel-resource-secretsmanager-certificate-v1"
description: |-
    Creates and manages a certificate in Selectel Secrets Manager using public API v1.
---

# selectel\_secretsmanager\_certificate_v1

Creates and manages a certificate in Selectel Secrets Manager using public API v1. For more information about certificates, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/secrets-manager/certificates/).

## Example Usage

### Text format

```hcl
resource "selectel_secretsmanager_certificate_v1" "certificate_1" {
  name          = "certificate"
  certificates  = [file("./_cert.pem")]
  private_key   = file("./_private_key.pem")
  project_id    = selectel_vpc_project_v2.project_1.id
}
```

### EOF

```hcl
resource "selectel_secretsmanager_certificate_v1" "certificate_1" {
  name         = "certificate"
  certificates = [
      <<-EOF
      -----BEGIN CERTIFICATE-----
      MIIDSzCCAjOgAwIBAgIULEumDHpDEHvQ1seZB9yRX9sCgoUwDQYJKoZIhvcNAQEL
      ...
      ----END CERTIFICATE-----
      EOF
  ]
  private_key  = <<-EOF
  -----BEGIN PRIVATE KEY-----
  MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCuk3SFn0AfAoxo
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

## Attributes Reference

* `dns_names` - Domain names for which the certificate is issued.

* `id` - Unique identifier of the certificate.

* `issued_by` - Information about the Certificate Authority (CA) which verified and signed the certificate.

* `serial` - Certificate serial number assigned by the Certificate Authority (CA) which issued the certificate.

* `validity` - Certificate validity in the RFC3339 timestamp format:

    * `not_before` - Effective date and time of the certificate.

    * `not_after` - Expiration date and time of the certificate.

* `version` - Certificate version.

## Import

You can import a certificate:

```shell
export INFRA_PROJECT_ID=<selectel_project_id>
terraform import selectel_secretsmanager_certificate_v1.certificate_1 <certificate_id>
```

where:

* `<selectel_project_id>` — Unique identifier of the associated project. To get the ID, in the [Control panel](https://my.selectel.ru/vpc/secrets-manager), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project. Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).

* `<certificate_id>` — Unique identifier of the certificate. To get the ID of the certificate, in the [Control panel](https://my.selectel.ru/vpc/secrets-manager/), go to **Cloud Platform** ⟶ **Secrets Manager** ⟶ the **Certificates** tab ⟶ in the certificate menu select **Copy UUID**.