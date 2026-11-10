# Set the project and pool in the provider configuration or the
# INFRA_PROJECT_ID and INFRA_REGION environment variables.
# Leave `image_id` and `backup_id` unset in the resource configuration after import.
# These fields are not restored. Setting either plans replacement of the volume.
# Set `volume_type` to the volume type name or leave it unset.
# A volume type UUID plans replacement of the volume.
import {
  to = selectel_compute_volume_v3.volume_1
  id = "<volume_uuid>"
}
