resource "selectel_craas_token_v1" "token_1" {
  project_id = selectel_vpc_project_v2.project_1.id
}

output "registry_username" {
  value = selectel_craas_token_v1.token_1.username
  sensitive = true
}

output "registry_token" {
  value = selectel_craas_token_v1.token_1.token
  sensitive = true
}
