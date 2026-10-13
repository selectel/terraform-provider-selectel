# Create an image from a URL
resource "selectel_compute_image_v2" "image_1" {
  project_id       = selectel_vpc_project_v2.project_1.id
  region           = "ru-6"
  name             = "ubuntu-24.04"
  container_format = "bare"
  disk_format      = "qcow2"
  image_source_url = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
  min_disk_gb      = 5

  properties = {
    os_distro = "ubuntu"
  }
}
