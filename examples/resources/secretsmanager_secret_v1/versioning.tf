# Update the value of an existing secret to upload a version without activating it.
resource "selectel_secretsmanager_secret_v1" "secret_1" {
  key        = "secret"
  value      = file("./secret.txt")
  project_id = selectel_vpc_project_v2.project_1.id
  is_current = false
}
