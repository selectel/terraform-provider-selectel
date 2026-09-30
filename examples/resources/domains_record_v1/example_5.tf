resource "selectel_domains_record_v1" "ns_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "NS"
  content   = "ns5.selectel.org"
  ttl       = 86400
}
