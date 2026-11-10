resource "selectel_domains_record_v1" "sshfp_record_1" {
  domain_id        = selectel_domains_domain_v1.main_domain.id
  name             = format("%s", selectel_domains_domain_v1.main_domain.name)
  type             = "SSHFP"
  ttl              = 60
  algorithm        = 1
  fingerprint_type = 1
  fingerprint      = "01AA"
}
