resource "selectel_mks_cluster_v1" "basic_cluster" {
  name                              = "cluster-1"
  project_id                        = selectel_vpc_project_v2.project_1.id
  region                            = "ru-7"
  kube_version                      = data.selectel_mks_kube_versions_v1.versions.latest_version
  cni_type                          = "CALICO"
  cluster_type                      = "BASIC"
  enable_patch_version_auto_upgrade = false
}
