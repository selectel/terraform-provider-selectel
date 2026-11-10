# Import the existing image by its UUID.
# `local_file_path`, `image_source_url`, and `properties` are not restored.
# Setting `local_file_path` or `image_source_url` after import does not replace the image.
# Set the project and pool containing the image.
export OS_DOMAIN_NAME=<domain_name>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>

terraform import selectel_compute_image_v2.image_1 <image_uuid>
