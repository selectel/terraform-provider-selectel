# Grant read-only access to the whole subnet of an NFS file storage
resource "selectel_filestorage_share_access_v2" "share_access_1" {
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-6"
  share_id     = selectel_filestorage_share_v2.share_1.id
  access_to    = "192.168.199.0/24"
  access_level = "ro"
}
