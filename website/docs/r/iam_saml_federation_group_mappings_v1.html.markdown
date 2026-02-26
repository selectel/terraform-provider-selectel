---
layout: "selectel"
page_title: "Selectel: selectel_iam_saml_federation_group_mappings_v1"
sidebar_current: "docs-selectel-resource-iam-saml-federation-group-mappings-v1"
description: |-
  Manages SAML Federation group mappings for Selectel products using public API v1.
---

# selectel\_iam\_saml\_federation\_group\_mappings\_v1

Manages SAML Federation group mappings for Selectel products using public API v1.
Selectel products support Identity and Access Management (IAM).
For more information about federations, see the [official Selectel documentation](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/federations/).

## Example Usage

```hcl
resource "selectel_iam_group_v1" "group_1" {
  name = "example-group"

  role {
    role_name = "reader"
    scope     = "account"
  }
}

resource "selectel_iam_saml_federation_v1" "federation_1" {
  name                  = "federation name"
  description           = "simple description"
  issuer                = "http://localhost:8080/realms/master"
  sso_url               = "http://localhost:8080/realms/master/protocol/saml"
  session_max_age_hours = 24
}

resource "selectel_iam_saml_federation_group_mappings_v1" "group_mappings_1" {
  federation_id = selectel_iam_saml_federation_v1.federation_1.id

  group_mapping {
    internal_group_id = selectel_iam_group_v1.group_1.id
    external_group_id = "external-group-1"
  }
}
```

## Argument Reference

* `federation_id` - (Required) Federation ID to manage group mappings for.

* `group_mapping` - (Required) One or more blocks defining mappings between internal IAM groups and external identity provider groups.

The `group_mapping` block supports:

* `internal_group_id` - (Required) Internal IAM group ID.

* `external_group_id` - (Required) External identity provider group ID.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - Resource ID. Equals to the `federation_id` value.

## Import

You can import SAML Federation group mappings:

```shell
export OS_DOMAIN_NAME=<account_id>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
terraform import selectel_iam_saml_federation_group_mappings_v1.group_mappings_1 <federation_id>
```

where:

* `<account_id>` — Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/en/control-panel-actions/account/registration/).

* `<username>` — Name of the service user. To get the name, in the [Control panel](https://my.selectel.ru/iam/users_management/users?type=service), go to **Identity & Access Management** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service Users](https://docs.selectel.ru/en/control-panel-actions/users-and-roles/user-types-and-roles/).

* `<password>` — Password of the service user.

* `<federation_id>` — Unique identifier of the federation, for example, `abc1bb378ac84e1234b869b77aadd2ab`. To get the federation ID, use either [Control Panel](https://my.selectel.ru/iam/federations) or [IAM API](https://developers.selectel.ru/docs/control-panel/iam/).

