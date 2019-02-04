output "project_id" {
  value = "${selectel_vpc_project_v2.webservice.id}"
}

output "webservice_floating_ip_ru1_1_id" {
  value = "${selectel_vpc_floatingip_v2.webservice_floating_ip_ru1_1.id}"
}

output "webservice_floating_ip_ru1_1_address" {
  value = "${selectel_vpc_floatingip_v2.webservice_floating_ip_ru1_1.floating_ip_address}"
}

output "webservice_floating_ip_ru1_2_id" {
  value = "${selectel_vpc_floatingip_v2.webservice_floating_ip_ru1_2.id}"
}

output "webservice_floating_ip_ru1_2_address" {
  value = "${selectel_vpc_floatingip_v2.webservice_floating_ip_ru1_2.floating_ip_address}"
}

output "webservice_floating_ip_ru2_1_id" {
  value = "${selectel_vpc_floatingip_v2.webservice_floating_ip_ru2_1.id}"
}

output "webservice_floating_ip_ru2_1_address" {
  value = "${selectel_vpc_floatingip_v2.webservice_floating_ip_ru2_1.floating_ip_address}"
}
