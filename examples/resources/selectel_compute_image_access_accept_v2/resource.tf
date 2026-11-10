# Accept access to a shared image
resource "selectel_compute_image_access_accept_v2" "image_access_accept_1" {
  project_id = selectel_vpc_project_v2.project_2.id
  region     = "ru-6"
  image_id   = selectel_compute_image_access_v2.image_access_1.image_id
  status     = "accepted"
}
