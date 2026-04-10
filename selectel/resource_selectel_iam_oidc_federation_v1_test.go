package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIAMV1OIDCFederationBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1OIDCFederationBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "name", "federation name"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "alias", "federation-alias"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "description", "simple description"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "issuer", "http://localhost:8080/realms/master"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "client_id", "my-client-id"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "client_secret", "my-client-secret"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "auth_url", "http://localhost:8080/realms/master/protocol/openid-connect/auth"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "token_url", "http://localhost:8080/realms/master/protocol/openid-connect/token"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "jwks_url", "http://localhost:8080/realms/master/protocol/openid-connect/certs"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "auto_users_creation", "true"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "enable_group_mappings", "true"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "session_max_age_hours", "24"),
				),
			},
		},
	})
}

func TestAccIAMV1OIDCFederationUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1OIDCFederationBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "name", "federation name"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "alias", "federation-alias"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "description", "simple description"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "issuer", "http://localhost:8080/realms/master"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "client_id", "my-client-id"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "client_secret", "my-client-secret"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "auth_url", "http://localhost:8080/realms/master/protocol/openid-connect/auth"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "token_url", "http://localhost:8080/realms/master/protocol/openid-connect/token"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "jwks_url", "http://localhost:8080/realms/master/protocol/openid-connect/certs"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "auto_users_creation", "true"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "enable_group_mappings", "true"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "session_max_age_hours", "24"),
				),
			},
			{
				Config: testAccIAMV1OIDCFederationUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "name", "federation name 2"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "alias", "federation-alias-2"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "description", "simple description 2"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "issuer", "http://localhost:8080/realms/master"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "client_id", "my-client-id"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "client_secret", "my-client-secret"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "auth_url", "http://localhost:8080/realms/master/protocol/openid-connect/auth"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "token_url", "http://localhost:8080/realms/master/protocol/openid-connect/token"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "jwks_url", "http://localhost:8080/realms/master/protocol/openid-connect/certs"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "auto_users_creation", "false"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "enable_group_mappings", "false"),
					resource.TestCheckResourceAttr("selectel_iam_oidc_federation_v1.federation_tf_acc_test_1", "session_max_age_hours", "24"),
				),
			},
		},
	})
}

func testAccIAMV1OIDCFederationBasic() string {
	return `
resource "selectel_iam_oidc_federation_v1" "federation_tf_acc_test_1" {
  name                  = "federation name"
  alias                 = "federation-alias"
  description           = "simple description"
  issuer                = "http://localhost:8080/realms/master"
  client_id             = "my-client-id"
  client_secret         = "my-client-secret"
  auth_url              = "http://localhost:8080/realms/master/protocol/openid-connect/auth"
  token_url             = "http://localhost:8080/realms/master/protocol/openid-connect/token"
  jwks_url              = "http://localhost:8080/realms/master/protocol/openid-connect/certs"
  auto_users_creation   = true
  enable_group_mappings = true
  session_max_age_hours = 24
}
`
}

func testAccIAMV1OIDCFederationUpdate() string {
	return `
resource "selectel_iam_oidc_federation_v1" "federation_tf_acc_test_1" {
  name                  = "federation name 2"
  alias                 = "federation-alias-2"
  description           = "simple description 2"
  issuer                = "http://localhost:8080/realms/master"
  client_id             = "my-client-id"
  client_secret         = "my-client-secret"
  auth_url              = "http://localhost:8080/realms/master/protocol/openid-connect/auth"
  token_url             = "http://localhost:8080/realms/master/protocol/openid-connect/token"
  jwks_url              = "http://localhost:8080/realms/master/protocol/openid-connect/certs"
  auto_users_creation   = false
  enable_group_mappings = false
  session_max_age_hours = 24
}
`
}
