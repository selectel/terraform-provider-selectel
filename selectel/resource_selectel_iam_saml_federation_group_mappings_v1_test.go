package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIAMV1SAMLFederationGroupMappingsBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1SAMLFederationGroupMappingsBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "federation_id"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "group_mapping.0.external_group_id", "external-group-1"),
				),
			},
		},
	})
}

func TestAccIAMV1SAMLFederationGroupMappingsUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1SAMLFederationGroupMappingsBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "federation_id"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "group_mapping.#", "1"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "group_mapping.0.external_group_id", "external-group-1"),
				),
			},
			{
				Config: testAccIAMV1SAMLFederationGroupMappingsUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "federation_id"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "group_mapping.#", "2"),
				),
			},
			{
				Config: testAccIAMV1SAMLFederationGroupMappingsBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "federation_id"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "group_mapping.#", "1"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_group_mappings_v1.group_mappings_tf_acc_test_1", "group_mapping.0.external_group_id", "external-group-1"),
				),
			},
		},
	})
}

func testAccIAMV1SAMLFederationGroupMappingsBasic() string {
	return `
resource "selectel_iam_group_v1" "group_tf_acc_test_1" {
  name = "test-group-mapping-1"
  role {
    role_name = "reader"
    scope     = "account"
  }
}

resource "selectel_iam_group_v1" "group_tf_acc_test_2" {
  name = "test-group-mapping-2"
  role {
    role_name = "reader"
    scope     = "account"
  }
}

resource "selectel_iam_saml_federation_v1" "federation_tf_acc_test_1" {
  name                  = "federation name"
  description           = "simple description"
  issuer                = "http://localhost:8080/realms/master"
  sso_url               = "http://localhost:8080/realms/master/protocol/saml"
  sign_authn_requests   = true
  force_authn           = true
  session_max_age_hours = 24
}

resource "selectel_iam_saml_federation_group_mappings_v1" "group_mappings_tf_acc_test_1" {
  federation_id = selectel_iam_saml_federation_v1.federation_tf_acc_test_1.id

  group_mapping {
    internal_group_id = selectel_iam_group_v1.group_tf_acc_test_1.id
    external_group_id = "external-group-1"
  }
}
`
}

func testAccIAMV1SAMLFederationGroupMappingsUpdate() string {
	return `
resource "selectel_iam_group_v1" "group_tf_acc_test_1" {
  name = "test-group-mapping-1"
  role {
    role_name = "reader"
    scope     = "account"
  }
}

resource "selectel_iam_group_v1" "group_tf_acc_test_2" {
  name = "test-group-mapping-2"
  role {
    role_name = "reader"
    scope     = "account"
  }
}

resource "selectel_iam_saml_federation_v1" "federation_tf_acc_test_1" {
  name                  = "federation name"
  description           = "simple description"
  issuer                = "http://localhost:8080/realms/master"
  sso_url               = "http://localhost:8080/realms/master/protocol/saml"
  sign_authn_requests   = true
  force_authn           = true
  session_max_age_hours = 24
}

resource "selectel_iam_saml_federation_group_mappings_v1" "group_mappings_tf_acc_test_1" {
  federation_id = selectel_iam_saml_federation_v1.federation_tf_acc_test_1.id

  group_mapping {
    internal_group_id = selectel_iam_group_v1.group_tf_acc_test_1.id
    external_group_id = "external-group-1"
  }

  group_mapping {
    internal_group_id = selectel_iam_group_v1.group_tf_acc_test_2.id
    external_group_id = "external-group-2"
  }
}
`
}
