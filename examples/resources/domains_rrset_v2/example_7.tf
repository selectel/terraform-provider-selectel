resource "selectel_domains_rrset_v2" "srv_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "_sip._tcp.example.com."
  type       = "SRV"
  ttl        = 120
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "10 20 30 example.org."
    # The content value is "<priority> <weight> <port> <target>"
  }
}
