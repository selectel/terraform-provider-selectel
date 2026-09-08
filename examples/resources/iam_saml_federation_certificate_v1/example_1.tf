resource "selectel_iam_saml_federation_certificate_v1" "certificate" {
  federation_id = selectel_iam_saml_federation_v1.federation_1.id
  name          = "certificate name"
  description   = "simple description"
  data          = file("${path.module}/federation_cert.crt")
}

