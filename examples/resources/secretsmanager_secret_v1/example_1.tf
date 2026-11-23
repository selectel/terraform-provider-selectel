resource "selectel_secretsmanager_secret_v1" "secret_1" {
  key        = "secret"
  value      = file("./secret.txt")
  project_id = selectel_vpc_project_v2.project_1.id
}
