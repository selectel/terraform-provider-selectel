---
layout: "selectel"
page_title: "Selectel: selectel_iam_oidc_federation_v1"
sidebar_current: "docs-selectel-resource-iam-oidc-federation-v1"
description: |-
  Creates and manages OIDC Federation for Selectel products using public API v1.
---

# selectel\_iam\_oidc\_federation\_v1

Manages OIDC Federation for Selectel products using public API v1.
Selectel products support Identity and Access Management (IAM).
For more information about federations, see the [official Selectel documentation](https://docs.selectel.ru/en/access-control/federations/).

## Example Usage

```hcl
resource "selectel_iam_oidc_federation_v1" "federation_1" {
  name                  = "Federation name"
  alias                 = "federation-alias"
  description           = "Federation description"
  issuer                = "https://idp.example.com/realms/master"
  client_id             = "my-client-id"
  client_secret         = "my-client-secret"
  auth_url              = "https://idp.example.com/realms/master/protocol/openid-connect/auth"
  token_url             = "https://idp.example.com/realms/master/protocol/openid-connect/token"
  jwks_url              = "https://idp.example.com/realms/master/protocol/openid-connect/certs"
  auto_users_creation   = true
  enable_group_mappings = true
  session_max_age_hours = 24
}
```

## Argument Reference

* `name` - (Required) Federation name.

* `alias` - (Optional) Federation alias.

* `description` - (Optional) Federation description.

* `issuer` - (Required) Unique identifier of the credential provider.

* `client_id` - (Required) Unique identifier of the client for OIDC authentication.

* `client_secret` - (Required, Sensitive) Client secret for OIDC authentication.

* `auth_url` - (Required) URL of the authorization endpoint used to authenticate users via the OIDC provider.

* `token_url` - (Required) URL of the token endpoint.

* `jwks_url` - (Required) URL of the JSON Web Key Set (JWKS) endpoint with certificates used for token verification.

* `auto_users_creation` - (Optional) Enables automatic user creation for this federation.

* `enable_group_mappings` - (Optional) Enables group mappings for this federation.

* `session_max_age_hours` - (Required) Session lifetime in hours.

## Attributes Reference

* `account_id` - Selectel account ID.

## Import

You can import a federation:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_iam_oidc_federation_v1.federation_1 <federation_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/service-users), go to **Account** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service Users](https://docs.selectel.ru/en/access-control/user-types/).

* `<password>` — Password of the service user.

* `<federation_id>` — Unique identifier of the federation, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the federation ID, in the [Control Panel](https://my.selectel.ru/iam/federations), go to **Account** → **Federations** → copy the ID under the federation name.
