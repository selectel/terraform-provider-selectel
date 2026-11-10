data "selectel_compute_image_ids_v2" "image_ids_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  name_regex = "^Ubuntu 24\\.04"
  sort       = "created_at:desc"
}
