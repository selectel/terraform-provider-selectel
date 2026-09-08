resource "selectel_mks_cluster_v1" "ha_cluster" {
  name         = "cluster-1"
  project_id   = selectel_vpc_project_v2.project_1.id
  region       = "ru-7"
  kube_version = data.selectel_mks_kube_versions_v1.versions.latest_version
  cni_type     = "CILIUM"
  cni_cilium_settings {
    envoy_daemonset = false
    hubble_relay    = true
  }
}
