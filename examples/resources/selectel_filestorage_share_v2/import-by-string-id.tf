# Set the project and pool in the provider configuration or the
# INFRA_PROJECT_ID and INFRA_REGION environment variables.
# Import does not restore `metadata`: after import, list in it only the keys you want Terraform to manage.
# A file storage accessible to all addresses of its subnet has a 0.0.0.0/0 access rule:
# import it as selectel_filestorage_share_access_v2.
import {
  to = selectel_filestorage_share_v2.share_1
  id = "<share_uuid>"
}
