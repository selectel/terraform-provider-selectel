data "selectel_mks_kubeconfig_v1" "kubeconfig" {
  cluster_id = selectel_mks_cluster_v1.cluster_1.id
  project_id = selectel_mks_cluster_v1.cluster_1.project_id
  region     = selectel_mks_cluster_v1.cluster_1.region
}

provider "kubernetes" {
  host                   = data.selectel_mks_kubeconfig_v1.kubeconfig.server
  client_certificate     = base64decode(data.selectel_mks_kubeconfig_v1.kubeconfig.client_cert)
  client_key             = base64decode(data.selectel_mks_kubeconfig_v1.kubeconfig.client_key)
  cluster_ca_certificate = base64decode(data.selectel_mks_kubeconfig_v1.kubeconfig.cluster_ca_cert)
}

output "kubeconfig" {
  value = data.selectel_mks_kubeconfig_v1.kubeconfig.raw_config
}
