# Create an empty network volume
resource "selectel_compute_volume_v3" "volume_1" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-6"
  size       = 10
}
