# Apply this change to an existing secret. Keep its last uploaded value unchanged.
# Keep ./secret.txt unchanged when pinning or rolling back.
resource "selectel_secretsmanager_secret_v1" "secret_1" {
  key             = "secret"
  value           = file("./secret.txt")
  project_id      = selectel_vpc_project_v2.project_1.id
  current_version = 1
}

# Changing value while pinned uploads an inactive version.
# Set current_version to that version number in a subsequent apply to activate it.
# Set current_version to 0 to automatically activate future value changes.
