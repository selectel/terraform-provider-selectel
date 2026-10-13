# Import the existing image access by the image UUID and the ID of the receiving project.
# Set the project that owns the image and its pool.
export OS_DOMAIN_NAME=<domain_name>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>

terraform import selectel_compute_image_access_v2.image_access_1 <image_uuid>/<member_id>
