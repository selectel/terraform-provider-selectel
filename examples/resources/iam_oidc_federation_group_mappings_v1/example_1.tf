resource "selectel_iam_group_v1" "group_1" {
  name = "Group name"

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
