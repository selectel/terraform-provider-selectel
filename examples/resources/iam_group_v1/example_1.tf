resource "selectel_iam_group_v1" "group_1" {
  name        = "My group"
  description = "My test group"
  role {
    role_name = "member"
    scope     = "account"
  }
}
