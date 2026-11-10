resource "selectel_domains_rrset_v2" "sshfp_rrset_1" {
  zone_id    = selectel_domains_zone_v2.zone_1.id
  name       = "example.com."
  type       = "SSHFP"
  ttl        = 60
  project_id = selectel_vpc_project_v2.project_1.id
  records {
    content = "1 1 7491973e5f8b39d5327cd4e08bc81b05f7710b49"
    # The content value is "<algorithm> <fingerprint_type> <fingerprint>"
  }
}
