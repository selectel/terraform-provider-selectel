# Import the existing access rule by the file storage UUID and the rule UUID.
# To get the rule UUID, run `openstack share access list <share_uuid>` and copy the `ID` field.
# Set the project and pool containing the file storage.
export OS_DOMAIN_NAME=<domain_name>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>

terraform import selectel_filestorage_share_access_v2.share_access_1 <share_uuid>/<access_uuid>
