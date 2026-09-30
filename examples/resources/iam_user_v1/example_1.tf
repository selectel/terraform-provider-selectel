resource "selectel_iam_user_v1" "user_1" {
  email       = "mail@example.com"
  role {
    role_name = "member"
    scope     = "account"
  }
}
