# Create an NFS file storage in a private subnet
resource "selectel_filestorage_share_v2" "share_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-6"
  name        = "data"
  share_proto = "NFS"
  share_type  = "basic.ru-6a"
  size        = 50
  network_id  = selectel_vpc_network_v2.network_1.id
  subnet_id   = selectel_vpc_network_subnet_v2.subnet_1.id
}

output "share_1_mount_path" {
  value = one([for location in selectel_filestorage_share_v2.share_1.export_locations : location.path if location.preferred])
}
