package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/secretsmanager-go/secretsmanagererrors"
)

func TestAccSecretsManagerCertificateV1ImportBasic(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")
	resourceName := "selectel_secretsmanager_certificate_v1.certificate_tf_acc_test_1"

	certificateName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSecretsManagerV1CertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsManagerCertificateV1WithoutProjectBasic(certificateName, projectID),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificates", "private_key"},
			},
		},
	})
}

func testAccSecretsManagerCertificateV1WithoutProjectBasic(certificateName, projectID string) string {
	return fmt.Sprintf(`
		resource "selectel_secretsmanager_certificate_v1" "certificate_tf_acc_test_1" {
			name = "%s"
			certificates = [
			<<-EOF
			-----BEGIN CERTIFICATE-----
			MIIDSzCCAjOgAwIBAgIULEumDHpDEHvQ1seZB9yRX9sCgoUwDQYJKoZIhvcNAQEL
			BQAwNTELMAkGA1UEBhMCUlUxEzARBgNVBAgMClNvbWUtU3RhdGUxETAPBgNVBAoM
			CFNlbGVjdGVsMB4XDTI0MDEwOTA4Mzc0M1oXDTM0MDEwNjA4Mzc0M1owNTELMAkG
			A1UEBhMCUlUxEzARBgNVBAgMClNvbWUtU3RhdGUxETAPBgNVBAoMCFNlbGVjdGVs
			MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArpN0hZ9AHwKMaPUQP4Z0
			4abCDxpKO2bJsdw1PxHOpkdw23dS2bH+wHWPspin5rK9i/wqg1fqKYikbukfBkdG
			WjHEpgHzjHuDER0dJ4iU8kD50kg64PaUHJ1fw2QfxmH7l/DNY+9poViqwJGpGCWp
			MsRw1OFQhLZKNhkNIgFugFesaBYJHdXqf7JAx+2y7AZBFniFl1PPs7Xtjn9j7m8i
			2WYc+1SgU8fI4uDhH+PxjIdNrwK5bC2xg68EXI0vSkyh6Ir74Va4FWW9tlsXpw3W
			d4NOorzmkDeSknbruhBHmbucmoh2oTcojziB2qRrlU8JcfjETJglZklLyzbXlk/N
			WwIDAQABo1MwUTAdBgNVHQ4EFgQU8RFMuHQ+rh0RYWYEmYozljJMrjQwHwYDVR0j
			BBgwFoAU8RFMuHQ+rh0RYWYEmYozljJMrjQwDwYDVR0TAQH/BAUwAwEB/zANBgkq
			hkiG9w0BAQsFAAOCAQEATn/WaWDnmUnYD4enM4U0HCQE6k+TodcPt3oMw+K0tfMP
			AKJkD+jJvqanH6ajZNWTgEmMoiEc6bv4D4/wsiSYSIjEQDOwTkVa1wYEXeXzYc5e
			GsnXXOusgR9+F5GFV8p8qDt4hozNtEycLbfN3gJURPqEJcwn7aJIVPeoWEOI5wO9
			banExY6twbb91OAdW8aTkD3qicsfRpDiYHVDKqgvEJpGCTWONeUnfcKy7ni4ahov
			PD3JcGkk8I+tbkM9gvxgKlXlGIHL3puskkusc5SxUSgDADLQwts5htT7TpOny7Dy
			peh6PHUaY/+beb4fwNtthbs1NvtVXFUVPlxJaPFW6A==
			-----END CERTIFICATE-----
			EOF
			]
			private_key = <<-EOF
			-----BEGIN PRIVATE KEY-----
			MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCuk3SFn0AfAoxo
			9RA/hnThpsIPGko7Zsmx3DU/Ec6mR3Dbd1LZsf7AdY+ymKfmsr2L/CqDV+opiKRu
			6R8GR0ZaMcSmAfOMe4MRHR0niJTyQPnSSDrg9pQcnV/DZB/GYfuX8M1j72mhWKrA
			kakYJakyxHDU4VCEtko2GQ0iAW6AV6xoFgkd1ep/skDH7bLsBkEWeIWXU8+zte2O
			f2PubyLZZhz7VKBTx8ji4OEf4/GMh02vArlsLbGDrwRcjS9KTKHoivvhVrgVZb22
			WxenDdZ3g06ivOaQN5KSduu6EEeZu5yaiHahNyiPOIHapGuVTwlx+MRMmCVmSUvL
			NteWT81bAgMBAAECggEAQCtITeN7BMsBhITr24XXSahrtXRy68G9CqkIU23+uSUS
			aUFDjWx9WQ39a2bsdIKn5KAkmlHC61BkLLZ45mxlgjq/70tRVAaEZ1J9yG3OXfuf
			OHm/VricOaZpMF+JxHh4q+FiBcVXXOzEGvOPpaYWOuh1FvLZD2cYASmVJ7ZCAV9d
			AB7YXmOQtnNtbe7BKa7aPHuK7zeyflpbCmaUBLJ7GR6UYV/xjJjp5clKHP0kt0OB
			E1gCveddwAVV7su/Oj1DEKI1w26fSBvmdVRf+pH4NddB1DYv2dr4scC/a2kTqZdn
			U+CUwG1Zd/LtxdCHQKDn36tIXYTKuX51WZ4jq4RcJQKBgQDHtcVbF6U9RTL8M3zu
			tMwyMTbbG6/2myupySYmOHzmV7XjqXCrbbMlSHQqBihE41XJo2ot8PETR6Q5mpWb
			BKbAYfUZVf93cbNIj29qESqp5adlvrwW3cbyDlMa81ehk+kdkPwUvlR2fvZP7PZr
			Om0eN6pFaK8ffCh6abbputwOlQKBgQDfyB4/0b8VTITkD6+DRkpbCOF0d+HJFJPr
			p3K7Gf06FSL5gqTT2SXdlQufVIee/X7QMefKLS90c0JjSpzqFIkFH0GB4nOtc9jQ
			sN3cEZjncWIVu1P493hBtx5Qb+oUVGBpGaDHk4hJgvUPx/t+NvNn1u/VwKHTHp5V
			4h0RbJygLwKBgAEt5ptyGUyyUunAWBWExcvqFHvYvwJCylA3Wt1Q6hPmIrHUd1Db
			1fn7Yow4+xXlDcWiDGd3C8VkX+jjK8z9iwqJyYu7wUVwS3G7PxouPcVBEOr95Fhy
			ONGHGiCHnVXb7L169LIeqZsFhujT6mSZtLk/9OZyBs61yftnEmhw7Qm9AoGBAMzG
			sD+QLP5NhjG31NEYykPhrYXJifhadz2mfguOrbWvz9Bo53HgfJD2qasETBKGP7w+
			XrAYhxtVuYNorIxbfEMOpgA3+8jWgKn/nxWZmMT5cVsXj7D8q7Pe4MOUlaxCxfKG
			/CSE8arrRltJke6eVEBKZC/C1ZJ+qz9F6XmfXPgLAoGAQ6MZASJ2qlYMA9xqmg/r
			/rMdjYdA9PBW3X0bcAa753ZltcZeV9Afrp+Mso8W5/c4fhUJ3mMgX1vGWnWcjnLP
			O8MUlX0Jx5czcb29DPSAdXjQ/pKXFyaIgjELxgb7APR0S7uQ9l7Cnm0S1Bd+ve7p
			Q5g85kFYklrYDOcltZ48JPs=
			-----END PRIVATE KEY-----
			EOF
		
			project_id = "%s"
		}
		`,
		certificateName,
		projectID,
	)
}

func testAccCheckSecretsManagerV1CertificateDestroy(s *terraform.State) error {
	smImportClient, diagErr := getSecretsManagerClientForAccImportTests(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get getSecretsManagerClientForAccImportTests for certificate import test")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_secretsmanager_certificate_v1" {
			continue
		}

		_, err := smImportClient.Certificates.Get(context.Background(), rs.Primary.ID)
		if !errors.Is(err, secretsmanagererrors.ErrNotFoundStatusText) {
			return errors.New("certificate still exists")
		}
	}

	return nil
}
