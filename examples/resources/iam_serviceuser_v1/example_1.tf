resource "selectel_iam_serviceuser_v1" "serviceuser_1" {
  name        = "username"
  password    = "password"
  role {
    role_name = "member"
    scope     = "account"
  }
  role {
    role_name = "iam_admin"
    scope     = "account"
  }
}
