resource "selectel_private_dns_zone_v1" "zone_1" {
	region     = "ru-1"
	project_id = selectel_vpc_project_v2.project_1.id
	domain     = "example.com."
	records {
		domain = "sub.example.com."
		type   = "A"
		values = [
			"192.168.0.2",
		]
	}
}
