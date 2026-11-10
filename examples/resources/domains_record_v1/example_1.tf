resource "selectel_domains_record_v1" "a_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "A"
  content   = "127.0.0.1"
  ttl       = 60
}
