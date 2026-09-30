data "selectel_global_router_quota_v1" "quota_1" {
  name        = "routers"
  scope       = "account_id"
  scope_value = "12345"
}
