resource "selectel_domains_record_v1" "alias_record_1" {
  domain_id = selectel_domains_domain_v1.main_domain.id
  name      = format("subc.%s", selectel_domains_domain_v1.main_domain.name)
  type      = "ALIAS"
  content   = format("%s", selectel_domains_domain_v1.main_domain.name)
  ttl       = 60
}
