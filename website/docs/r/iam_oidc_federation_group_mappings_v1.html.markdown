---
layout: "selectel"
page_title: "Selectel: selectel_iam_oidc_federation_group_mappings_v1"
sidebar_current: "docs-selectel-resource-iam-oidc-federation-group-mappings-v1"
description: |-
  Manages OIDC Federation group mappings for Selectel products using public API v1.
---

# selectel\_iam\_oidc\_federation\_group\_mappings\_v1

Manages OIDC federation group mappings for Selectel products using public API v1.
Selectel products support Identity and Access Management (IAM).
For more information about federations, see the [official Selectel documentation](https://docs.selectel.ru/access-control/federations/).

## Example Usage

```hcl
resource "selectel_iam_group_v1" "group_1" {
  name = "example-group"

  role {
    role_name = "reader"
    scope     = "account"
  }
}

resource "selectel_iam_oidc_federation_v1" "federation_1" {
  name                  = "Federation name"
  description           = "Federation description"
  issuer                = "https://idp.example.com/realms/master"
  client_id             = "my-client-id"
  client_secret         = "my-client-secret"
  auth_url              = "https://idp.example.com/realms/master/protocol/openid-connect/auth"
  token_url             = "https://idp.example.com/realms/master/protocol/openid-connect/token"
  jwks_url              = "https://idp.example.com/realms/master/protocol/openid-connect/certs"
  session_max_age_hours = 24
}

resource "selectel_iam_oidc_federation_group_mappings_v1" "group_mappings_1" {
  federation_id = selectel_iam_oidc_federation_v1.federation_1.id

  group_mapping {
    internal_group_id = selectel_iam_group_v1.group_1.id
    external_group_id = "external-group-1"
  }
}
```

## Argument Reference

* `federation_id` - (Required) Federation ID to manage group mappings for.

* `group_mapping` - (Required) Defines mappings between internal IAM groups and external identity provider groups. You can add multiple mappings – each mapping in a separate block.

    * `internal_group_id` - (Required) Internal IAM group ID.

    * `external_group_id` - (Required) External identity provider group ID.

## Attributes Reference

* `id` - Resource ID. Equals the `federation_id` value.

## Import

You can import OIDC Federation group mappings:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_iam_oidc_federation_group_mappings_v1.group_mappings_1 <federation_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/service-users), go to **Account** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service Users](https://docs.selectel.ru/access-control/user-types/).

* `<password>` — Password of the service user.

* `<federation_id>` — Unique identifier of the federation, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the federation ID, use either [Control Panel](https://my.selectel.ru/iam/federations) or [Federations API](https://docs.selectel.ru/api/federations/).
