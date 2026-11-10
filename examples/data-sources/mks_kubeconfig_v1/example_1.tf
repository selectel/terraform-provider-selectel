data "selectel_mks_kubeconfig_v1" "kubeconfig" {
  cluster_id = selectel_mks_cluster_v1.cluster_1.id
  project_id = selectel_mks_cluster_v1.cluster_1.project_id
  region     = selectel_mks_cluster_v1.cluster_1.region
}

output "kubeconfig" {
  value = data.selectel_mks_kubeconfig_v1.kubeconfig.raw_config
}
