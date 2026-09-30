# Configure the Selectel provider

provider "selectel" {
  domain_name = "123456"
  username    = "user"
  password    = "password"
  auth_region = "pool"
  auth_url = "https://cloud.api.selcloud.ru/identity/v3/"
}
