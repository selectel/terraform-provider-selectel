resource "selectel_iam_s3_credentials_v1" "s3_credentials_1" {
  user_id    = selectel_iam_serviceuser_v1.serviceuser_1.id
  project_id = selectel_vpc_project_v2.project_1.id
  name       = "S3Credentials"
}
