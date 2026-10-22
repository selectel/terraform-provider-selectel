# Import the access by the image UUID and the ID of the project that received access.
# With `<image_uuid>` alone, Terraform takes the project from INFRA_PROJECT_ID.
# Set the pool of the image.
export OS_DOMAIN_NAME=<domain_name>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_REGION=<region>

terraform import selectel_compute_image_access_accept_v2.image_access_accept_1 <image_uuid>/<project_id>
