# Create an SSD Fast v2 network volume with custom IOPS
resource "selectel_compute_volume_v3" "volume_1" {
  project_id        = selectel_vpc_project_v2.project_1.id
  region            = "ru-6"
  size              = 10
  availability_zone = "ru-6a"
  volume_type       = data.selectel_compute_volume_type_v3.volume_type_1.name

  metadata = {
    total_iops_sec = "30000"
  }
}
