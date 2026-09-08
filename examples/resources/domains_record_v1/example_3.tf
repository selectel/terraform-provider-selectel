resource "selectel_domains_record_v1" "txt_record_1" {
  domain_id = selectel_domains_domain_v1.domain_1.id
  name      = "example.com"
  type      = "TXT"
  content   = "hello, world!"
  ttl       = 60
}
