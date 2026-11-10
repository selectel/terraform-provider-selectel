resource "selectel_domains_record_v1" "srv_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "SRV"
  ttl       = 120
  priority  = 10
  weight    = 20
  target    = "example.com"
  port      = 100
}
