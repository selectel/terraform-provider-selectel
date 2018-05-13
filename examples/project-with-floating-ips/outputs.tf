output "project_id" {
  value = "${selvpc_resell_project_v2.webservice.id}"
}

output "webservice_floating_ip_ru1_1_id" {
  value = "${selvpc_resell_floatingip_v2.webservice_floating_ip_ru1_1.id}"
}

output "webservice_floating_ip_ru1_1_address" {
  value = "${selvpc_resell_floatingip_v2.webservice_floating_ip_ru1_1.floating_ip_address}"
}

output "webservice_floating_ip_ru1_2_id" {
  value = "${selvpc_resell_floatingip_v2.webservice_floating_ip_ru1_2.id}"
}

output "webservice_floating_ip_ru1_2_address" {
  value = "${selvpc_resell_floatingip_v2.webservice_floating_ip_ru1_2.floating_ip_address}"
}

output "webservice_floating_ip_ru2_1_id" {
  value = "${selvpc_resell_floatingip_v2.webservice_floating_ip_ru2_1.id}"
}

output "webservice_floating_ip_ru2_1_address" {
  value = "${selvpc_resell_floatingip_v2.webservice_floating_ip_ru2_1.floating_ip_address}"
}
