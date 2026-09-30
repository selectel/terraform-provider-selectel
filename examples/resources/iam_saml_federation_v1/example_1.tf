resource "selectel_iam_saml_federation_v1" "federation_1" {
  name                  = "Federation name"
  alias                 = "federation-alias"
  description           = "Federation description"
  issuer                = "http://localhost:8080/realms/master"
  sso_url               = "http://localhost:8080/realms/master/protocol/saml"
  sign_authn_requests   = true
  force_authn           = true
  auto_users_creation   = true
  enable_group_mappings = true
  session_max_age_hours = 24
}
