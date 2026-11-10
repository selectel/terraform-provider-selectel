resource "selectel_craas_token_v2" "token_1" {
  project_id     = selectel_vpc_project_v2.project_1.id
  name           = "terraform-token-270295000"
  mode_rw        = true
  all_registries = true
  registry_ids   = []
  is_set         = true
  expires_at     = "2029-01-01T00:00:00Z"
}

output "registry_username" {
  value     = selectel_craas_token_v2.token_1.username
  sensitive = true
}

output "registry_token" {
  value     = selectel_craas_token_v2.token_1.token
  sensitive = true
}
