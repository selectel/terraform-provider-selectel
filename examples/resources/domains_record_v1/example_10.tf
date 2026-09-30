resource "selectel_domains_record_v1" "caa_record_1" {
  domain_id = selectel_domains_domain_v1.main_domain.id
  name      = format("caa.%s", selectel_domains_domain_v1.main_domain.name)
  type      = "CAA"
  ttl       = 60
  tag       = "issue"
  flag      = 128
  value     = "letsencrypt.com"
}
