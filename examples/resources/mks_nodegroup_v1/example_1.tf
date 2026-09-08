resource "selectel_mks_nodegroup_v1" "nodegroup_1" {
  cluster_id        = selectel_mks_cluster_v1.cluster_1.id
  project_id        = selectel_mks_cluster_v1.cluster_1.project_id
  region            = selectel_mks_cluster_v1.cluster_1.region
  availability_zone = "ru-7a"
  nodes_count       = 3
  cpus              = 2
  ram_mb            = 4096
  volume_gb         = 20
  volume_type       = "fast.ru-7a"

  install_nvidia_device_plugin = false
  preemptible                  = false

  labels            = {
    "label-key0": "label-value0",
    "label-key1": "label-value1",
    "label-key2": "label-value2",
  }
  taints {
    key    = "test-key-0"
    value  = "test-value-0"
    effect = "NoSchedule"
  }
  taints {
    key    = "test-key-1"
    value  = "test-value-1"
    effect = "NoExecute"
  }
  taints {
    key    = "test-key-2"
    value  = "test-value-2"
    effect = "PreferNoSchedule"
  }
}
