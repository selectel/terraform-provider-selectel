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
