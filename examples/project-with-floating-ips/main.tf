resource "selvpc_resell_project_v2" "webservice" {
  name   = "webservice"
  quotas = [
    {
      resource_name = "compute_cores"
      resource_quotas = [
        {
          region = "ru-1"
          zone = "ru-1a"
          value = 4
        },
        {
          region = "ru-2"
          zone = "ru-2a"
          value = 6
        },
      ]
    },
    {
      resource_name   = "compute_ram"
      resource_quotas = [
        {
          region = "ru-1"
          zone = "ru-1a"
          value = 10240
        },
        {
          region = "ru-2"
          zone = "ru-2a"
          value = 8192
        },
      ]
    },
    {
      resource_name   = "volume_gigabytes_fast"
      resource_quotas = [
        {
          region = "ru-1"
          zone = "ru-1a"
          value = 10
        },
        {
          region = "ru-2"
          zone = "ru-2a"
          value = 8
        }
      ]
    }
  ]
}

resource "selvpc_resell_floatingip_v2" "webservice_floating_ip_ru1_1" {
  project_id = "${selvpc_resell_project_v2.webservice.id}"
  region     = "ru-1"
  depends_on = ["selvpc_resell_project_v2.webservice"]
}

resource "selvpc_resell_floatingip_v2" "webservice_floating_ip_ru1_2" {
  project_id = "${selvpc_resell_project_v2.webservice.id}"
  region     = "ru-1"
  depends_on = ["selvpc_resell_project_v2.webservice"]
}

resource "selvpc_resell_floatingip_v2" "webservice_floating_ip_ru2_1" {
  project_id = "${selvpc_resell_project_v2.webservice.id}"
  region     = "ru-2"
  depends_on = ["selvpc_resell_project_v2.webservice"]
}
