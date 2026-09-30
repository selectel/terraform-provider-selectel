resource "selectel_vpc_license_v2" "license_windows_2016_standard" {
  project_id = selectel_vpc_project_v2.project_1.id
  region     = "ru-2"
  type       = "license_windows_2012_standard"
}
