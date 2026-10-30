# Create a file storage in a pool segment from the list of availability zones
data "selectel_filestorage_availability_zones_v2" "availability_zones_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
}

resource "selectel_filestorage_share_v2" "share_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-6"
  name        = "data"
  share_proto = "NFS"
  share_type  = "basic.${data.selectel_filestorage_availability_zones_v2.availability_zones_1.names[0]}"
  size        = 50
  network_id  = selectel_vpc_network_v2.network_1.id
  subnet_id   = selectel_vpc_network_subnet_v2.subnet_1.id
}
