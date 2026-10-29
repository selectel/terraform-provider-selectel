# Import the existing file storage by its UUID.
# Import does not restore `metadata`: after import, list in it only the keys you want Terraform to manage.
# A file storage accessible to all addresses of its subnet has a 0.0.0.0/0 access rule:
# import it as selectel_filestorage_share_access_v2.
# Set the project and pool containing the file storage.
export OS_DOMAIN_NAME=<domain_name>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>

terraform import selectel_filestorage_share_v2.share_1 <share_uuid>
