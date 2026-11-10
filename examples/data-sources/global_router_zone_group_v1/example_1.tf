data "selectel_global_router_zone_v1" "zone_1" {
  name    = "ru-3"
  service = "vpc"
}

data "selectel_global_router_zone_group_v1" "zone_group_1" {
  name = data.selectel_global_router_zone_v1.zone_1.groups[0].name
}
