resource "selectel_domains_record_v1" "mx_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "MX"
  content   = "mail.example.org"
  ttl       = 60
  priority  = 10
}
