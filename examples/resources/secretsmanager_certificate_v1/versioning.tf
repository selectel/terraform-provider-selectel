# Apply this change to an existing certificate. Keep its last uploaded PEM unchanged.
# Version 1 also exists immediately after creation.
resource "selectel_secretsmanager_certificate_v1" "certificate_1" {
  name            = "certificate"
  certificates    = [file("./_cert.pem")]
  private_key     = file("./_private_key.pem")
  project_id      = selectel_vpc_project_v2.project_1.id
  current_version = 1
}

# Changing PEM while pinned uploads an inactive version.
# Set current_version to that version number in a subsequent apply to activate it.
# Setting current_version to 0 leaves the active version unchanged.
# Future content changes then activate their new versions automatically.
