resource "selectel_secretsmanager_certificate_v1" "certificate_1" {
  name         = "certificate"
  certificates = [
      <<-EOF
      -----BEGIN CERTIFICATE-----
      MIIDSzCCAjOgAwIBAgIULEumDHpDEHvQ1seZB9yRX9sCgoUwDQYJKoZIhvcNAQEL
      ...
      ----END CERTIFICATE-----
      EOF
  ]
  private_key  = <<-EOF
  -----BEGIN PRIVATE KEY-----
  MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCuk3SFn0AfAoxo
  ...
  -----END PRIVATE KEY-----
  EOF
  project_id   = selectel_vpc_project_v2.project_1.id
}
