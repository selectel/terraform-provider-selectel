# Import the existing network volume by its UUID.
# Leave `image_id` and `backup_id` unset in the resource configuration after import.
# These fields are not restored. Setting either plans replacement of the volume.
# Set `volume_type` to the volume type name or leave it unset.
# A volume type UUID plans replacement of the volume.
# Set the project and pool containing the volume.
export OS_DOMAIN_NAME=<domain_name>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>

terraform import selectel_compute_volume_v3.volume_1 <volume_uuid>
