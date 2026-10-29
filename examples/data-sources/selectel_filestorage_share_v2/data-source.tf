data "selectel_filestorage_share_v2" "share_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name       = "data"
}
