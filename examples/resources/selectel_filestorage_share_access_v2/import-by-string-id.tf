# To get the rule UUID, run `openstack share access list <share_uuid>` and copy the `ID` field.
# Set the project and pool in the provider configuration or the
# INFRA_PROJECT_ID and INFRA_REGION environment variables.
import {
  to = selectel_filestorage_share_access_v2.share_access_1
  id = "<share_uuid>/<access_uuid>"
}
