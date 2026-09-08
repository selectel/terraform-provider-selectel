data "selectel_domains_rrset_v2" "rrset_1" {
  name       = "example.com."
  type       = "A"
  zone_id    = selectel_domains_zone_v2.zone_1.id
  project_id = selectel_vpc_project_v2.project_1.id
}
