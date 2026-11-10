data "selectel_compute_image_v2" "image_1" {
  project_id  = selectel_vpc_project_v2.project_1.id
  region      = "ru-6"
  name        = "Ubuntu 24.04 LTS 64-bit"
  visibility  = "public"
  most_recent = true
}
