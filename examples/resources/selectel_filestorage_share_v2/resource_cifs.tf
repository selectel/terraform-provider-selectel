# Create a CIFS file storage with a fixed IP address
resource "selectel_filestorage_share_v2" "share_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-6"
  name        = "windows-data"
  share_proto = "CIFS"
  share_type  = "basic.ru-6a"
  size        = 50
  network_id  = selectel_vpc_network_v2.network_1.id
  subnet_id   = selectel_vpc_network_subnet_v2.subnet_1.id
  ip          = "192.168.199.50"

  metadata = {
    owner = "team"
  }
}
