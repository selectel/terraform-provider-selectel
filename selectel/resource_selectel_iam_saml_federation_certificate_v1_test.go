package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var cert = `-----BEGIN CERTIFICATE-----\nMIICmzCCAYMCBgGI6ANFczANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMjMwNjIzMTEyNjQ4WhcNMzMwNjIzMTEyODI4WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC04rOaDpre/MucE3HXVCnAnpqIqQOeMn696AW2FATnI26x1BsxVAGjcrheAOIu+CxC28m48Ah4+SiTEk/u2X/WbGTd/1GZooz37cge0AWMQGyh8ysZRd6q06kg4QGD1iUtdQyHioMbSr9pPne2QQgSX5/gM9XDuA6dpG9Yv0PIPLFlk3BIUL1qEfUiYbDlrunkN/y4XromJaJPpgXKWraH194bqcgXGQLrCqicKwsRBoQJHg3ODWHjHFOwYODJ1XBsRcAue4J88PKiPV1tZNPVczMptrkqGBYTgOYGjKXGe5EH50RJE4/3Ynurz2s34DSDVJhJOYtGwpfeSuU3i3mVAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAGAweCuWJmJXMUdRtgoFIiu6BGotDX5sA/VOm4CRsEXV7/qnBagrAPkRz86KGm4lOPL0X+I13JQh4/OB1gxnPN+BXhNtCWCoj1wA3/BWjs1ow/gaVXzwdy+1mbc/sUBudsLq2Yqs54GgeYsTBKMVpSLKiRg1NebEFlqFmG2hjPzYg1QHL4VBusMQgqt7TTnOfGtdT3Ss9TKGRQ+iwfNL0BtSAKaTRdhNVU4lDYUs788Kw5od/uJj0wTICKO5/PrkX7Uy42+fyU+4SvJynPOy+M+z+s08JC9+eYXixfeeFG1nNWR+DIKXcXaSwNQW+8RweGbOJxQ2BoUKtl0NCHrvxJw=\n-----END CERTIFICATE-----`

func TestAccIAMV1SAMLFederationCertificateBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1SAMLFederationCertificateBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "name", "cert"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "description", "simple description"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "data"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "not_before"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "not_after"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "fingerprint"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMV1SAMLFederationCertificateUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1SAMLFederationCertificateBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "name", "cert"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "description", "simple description"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "data"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "not_before"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "not_after"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "fingerprint"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccIAMV1SAMLFederationCertificateUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "name", "cert 2"),
					resource.TestCheckResourceAttr("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "description", "simple description 2"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "data"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "not_before"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "not_after"),
					resource.TestCheckResourceAttrSet("selectel_iam_saml_federation_certificate_v1.certificate_tf_acc_test_1", "fingerprint"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIAMV1SAMLFederationCertificateBasic() string {
	return fmt.Sprintf(`
resource "selectel_iam_saml_federation_v1" "federation_tf_acc_test_1" {
  name                  = "federation name"
  description           = "simple description"
  issuer                = "http://localhost:8080/realms/master"
  sso_url               = "http://localhost:8080/realms/master/protocol/saml"
  sign_authn_requests   = true
  force_authn           = true
  session_max_age_hours = 24
}

resource "selectel_iam_saml_federation_certificate_v1" "certificate_tf_acc_test_1" {
  federation_id = selectel_iam_saml_federation_v1.federation_tf_acc_test_1.id
  name          = "cert"
  description   = "simple description"
  data          = "%s"
}
`, cert)
}

func testAccIAMV1SAMLFederationCertificateUpdate() string {
	return fmt.Sprintf(`
resource "selectel_iam_saml_federation_v1" "federation_tf_acc_test_1" {
  name                  = "federation name"
  description           = "simple description"
  issuer                = "http://localhost:8080/realms/master"
  sso_url               = "http://localhost:8080/realms/master/protocol/saml"
  sign_authn_requests   = true
  force_authn           = true
  session_max_age_hours = 24
}

resource "selectel_iam_saml_federation_certificate_v1" "certificate_tf_acc_test_1" {
  federation_id = selectel_iam_saml_federation_v1.federation_tf_acc_test_1.id
  name          = "cert 2"
  description   = "simple description 2"
  data          = "%s"
}
`, cert)
}
