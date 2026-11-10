# Set the pool in the provider configuration or the INFRA_REGION environment variable.
# With `<image_uuid>` alone, Terraform takes the project from the provider configuration
# or the INFRA_PROJECT_ID environment variable.
import {
  to = selectel_compute_image_access_accept_v2.image_access_accept_1
  id = "<image_uuid>/<project_id>"
}
