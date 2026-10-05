resource "selectel_vpc_rbac_policy_v2" "rbac_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-6"
  action            = "access_as_shared"
  object_type       = "network"
  object_id         = selectel_vpc_network_v2.network_1.id
  target_project_id = selectel_vpc_project_v2.project_2.id
}
