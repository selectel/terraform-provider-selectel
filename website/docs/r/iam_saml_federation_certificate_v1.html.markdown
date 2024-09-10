---
layout: "selectel"
page_title: "Selectel: selectel_iam_saml_federation_certificate_v1"
sidebar_current: "docs-selectel-resource-iam-saml-federation-certificate-v1"
description: |-
  Creates and manages SAML Federation Certificates for Selectel products using public API v1.
---

# selectel\_iam\_saml\_federation\_certificate\_v1

Manages SAML Federation Certificates for Selectel products using public API v1.
Selectel products support Identity and Access Management (IAM).
For more information about Federation Certificates, see the [official Selectel documentation](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/federations/certificates/).

## Example Usage

```hcl
resource "selectel_iam_saml_federation_certificate_v1" "certificate" {
  federation_id = selectel_iam_saml_federation_v1.federation_1.id
  name          = "certificate name"
  description   = "simple description"
  data          = file("${path.module}/federation_cert.crt")
}

```

## Argument Reference

* `federation_id` - (Required) Unique identifier of the federation.

* `name` - (Required) Certificate name.

* `description` - (Optional) Certificate description.

* `data` - (Required) Certificate data. Must begin with `-----BEGIN CERTIFICATE-----` and end with `-----END CERTIFICATE-----`.

## Attributes Reference

* `account_id` - Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `not_before` - Issue date of the certificate.

* `not_after` - Expiration date of the certificate.

* `fingerprint` - Fingerprint of the certificate.

## Import

You can import a certificate:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export OS_SAML_FEDERATION_ID=<federation_id>
terraform import selectel_iam_saml_federation_certificate_v1.certificate_1 <certificate_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service Users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<federation_id>` — Unique identifier of the associated federation, for which the certificate is issued, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the federation ID, use either [Control Panel](https://my.selectel.ru/iam/federations) or [IAM API](https://developers.selectel.ru/docs/control-panel/iam/).

* `<certificate_id>` — Unique identifier of the certificate.
