resource "selectel_domains_rrset_v2" "mx_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "MX"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "10 mail.example.org."
    # The content value is "<priority> <host>"
  }
}
