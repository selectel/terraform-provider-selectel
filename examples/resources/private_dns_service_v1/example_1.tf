resource "selectel_private_dns_service_v1" "service_1" {
	region     = "ru-1"
	project_id = selectel_vpc_project_v2.project_1.id
	network_id = openstack_networking_network_v2.network_1.id
	is_recursor_enabled = true

	depends_on = [
		openstack_networking_network_v2.network_1
    ]
}
