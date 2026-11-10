resource "selectel_secretsmanager_certificate_v1" "certificate_1" {
  name          = "certificate"
  certificates  = [file("./_cert.pem")]
  private_key   = file("./_private_key.pem")
  project_id    = selectel_vpc_project_v2.project_1.id
}
