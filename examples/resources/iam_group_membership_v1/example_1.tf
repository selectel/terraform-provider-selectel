resource "selectel_iam_group_membership_v1" "group_membership_1" {
  group_id = selectel_iam_group_v1.group_1.id
  
  user_ids = [
    selectel_iam_user_v1.user_1.keystone_id,
    selectel_iam_serviceuser_v1.serviceuser_1.id
  ]
}
