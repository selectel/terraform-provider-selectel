data "selectel_dedicated_configuration_v1" "server_config" {
  project_id  = selectel_vpc_project_v2.project_1.id
  deep_filter = <<EOT
    {
        "gpu": {
           "count": 1
        },
        "state": "Active",
    }
  EOT
}
