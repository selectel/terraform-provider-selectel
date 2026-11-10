resource "selectel_domains_rrset_v2" "a_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "A"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "127.0.0.1"
    # The content value is "<ipv4_address>"
  }
}
