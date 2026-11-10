resource "selectel_domains_rrset_v2" "ns_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "subdomain.example.com."
  type       = "NS"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "a.ns.selectel.ru."
    # The content value is "<name_server>"
  }
}
