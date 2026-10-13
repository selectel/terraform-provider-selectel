# Upload an image from a local file
resource "selectel_compute_image_v2" "image_1" {
  project_id       = selectel_vpc_project_v2.project_1.id
  region           = "ru-6"
  name             = "image-1"
  container_format = "bare"
  disk_format      = "qcow2"
  local_file_path  = "/var/lib/images/disk.qcow2"
}
