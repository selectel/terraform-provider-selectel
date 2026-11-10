resource "selectel_domains_rrset_v2" "aaaa_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "AAAA"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "2400:cb00:2049:1::a29f:1804"
    # The content value is "<ipv6_address>"
  }
}
