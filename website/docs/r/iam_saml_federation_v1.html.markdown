---
layout: "selectel"
page_title: "Selectel: selectel_iam_saml_federation_v1"
sidebar_current: "docs-selectel-resource-iam-saml-federation-v1"
description: |-
  Creates and manages SAML Federation for Selectel products using public API v1.
---

# selectel\_iam\_saml\_federation\_v1

Manages SAML Federation for Selectel products using public API v1.
Selectel products support Identity and Access Management (IAM).
For more information about federations, see the [official Selectel documentation](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/federations/).

## Example Usage

```hcl
resource "selectel_iam_saml_federation_v1" "federation_1" {
  name                  = "federation name"
  description           = "simple description"
  issuer                = "http://localhost:8080/realms/master"
  sso_url               = "http://localhost:8080/realms/master/protocol/saml"
  session_max_age_hours = 24
}
```

## Argument Reference

* `name` - (Required) Federation name.

* `description` - (Optional) Federation description.

* `issuer` - (Required) Identifier of the credential provider.

* `sso_url` - (Required) Link to the credential provider login page.

* `sign_authn_requests` - (Optional) Enables signing of authentication requests.

* `force_authn` - (Optional) Requires users to authenticate via SSO every time they log in.

* `session_max_age_hours` - (Required) Session lifetime.

## Attributes Reference

* `account_id` - Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

## Import

You can import a federation:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_iam_saml_federation_v1.federation_1 <federation_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service Users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<federation_id>` — Unique identifier of the federation, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the federation ID, use either [Control Panel](https://my.selectel.ru/iam/federations) or [IAM API](https://developers.selectel.ru/docs/control-panel/iam/).
