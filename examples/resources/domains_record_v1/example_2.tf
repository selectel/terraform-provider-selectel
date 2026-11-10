resource "selectel_domains_record_v1" "aaaa_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "AAAA"
  content   = "2400:cb00:2049:1::a29f:1804"
  ttl       = 60
}
