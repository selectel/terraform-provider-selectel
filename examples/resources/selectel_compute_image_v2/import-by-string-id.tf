# Set the project and pool in the provider configuration or the
# INFRA_PROJECT_ID and INFRA_REGION environment variables.
# `local_file_path`, `image_source_url`, and `properties` are not restored.
# Setting `local_file_path` or `image_source_url` after import does not replace the image.
import {
  to = selectel_compute_image_v2.image_1
  id = "<image_uuid>"
}
