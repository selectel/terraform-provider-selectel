resource "selectel_craas_registry_v1" "registry_1" {
  name       = "my-first-registry"
  project_id = selectel_vpc_project_v2.project_1.id
}
