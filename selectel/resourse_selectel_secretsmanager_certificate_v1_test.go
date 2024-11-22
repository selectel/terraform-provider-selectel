package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccSecretsManagerCertificateV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	certificateName := acctest.RandomWithPrefix("tf-acc")
	newCertificateName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsManagerCertificateV1BasicConfig(projectName, certificateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("selectel_secretsmanager_certificate_v1.certificate_tf_acc_test_1", "name", certificateName),
					resource.TestCheckResourceAttr("selectel_secretsmanager_certificate_v1.certificate_tf_acc_test_1", "version", "1"),
				),
			},
			{
				Config: testAccSecretsManagerCertificateV1UpdateConfig(projectName, newCertificateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("selectel_secretsmanager_certificate_v1.certificate_tf_acc_test_1", "name", newCertificateName),
					resource.TestCheckResourceAttr("selectel_secretsmanager_certificate_v1.certificate_tf_acc_test_1", "version", "2"),
				),
			},
		},
	})
}

func testAccSecretsManagerCertificateV1BasicConfig(projectName, certificateName string) string {
	return fmt.Sprintf(`
		resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
			name = "%s"
		}

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
		
			project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
		}
		`,
		projectName,
		certificateName,
	)
}

func testAccSecretsManagerCertificateV1UpdateConfig(projectName, newCertificateName string) string {
	return fmt.Sprintf(`
		resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
			name = "%s"
		}

		resource "selectel_secretsmanager_certificate_v1" "certificate_tf_acc_test_1" {
			name = "%s"
			certificates = [
			<<-EOF
			-----BEGIN CERTIFICATE-----
			MIIDazCCAlOgAwIBAgIUNUDIWJPP3qwVPsd8YZ+MfCW8o00wDQYJKoZIhvcNAQEL
			BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
			GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yNDAyMjYxMjQ0MjlaFw0zNDAy
			MjMxMjQ0MjlaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
			HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
			AQUAA4IBDwAwggEKAoIBAQCnUfnSg19ofGhnyYvG7I9Bt+lnaFbtmn7ZTjR6ASeq
			cFOTfqL79/D5R5en+j+dJa0FsoQfJ1yDjxHvANwXBuhyGb8xLNdy3y9TlzP7BLGN
			9Zt2doJmSokuILIczpNxEnY0pSPXyNEJhH6xwDIH8I4nPaE76M1ZZ/n5pNxNNYkZ
			UH+7UIk2m8LsNFtMpUmcelVTDN827DicncWvJw46EqxiGTlLqnXzyb7pyh7BIcj5
			GyYyRfGBsOcc1geL+ci1e3XIkDi89kXe4qlofSWx+2IwHYRFKLga6kQvHM7cd2XJ
			nX23Pal6hWVSuvxaaeHatcxBcnsv5L+UvZUyKsTS7qBTAgMBAAGjUzBRMB0GA1Ud
			DgQWBBRfeV2sCfbsvm+GQY3mAOY8wBE9KzAfBgNVHSMEGDAWgBRfeV2sCfbsvm+G
			QY3mAOY8wBE9KzAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCB
			07TyMwYSrbvjx4VBoe0v/oZ3Q0XFhAUBnNOEYd+AXJO6AvBXS1OUPi+yOhI6lU98
			iKlw6u8khK9+zDdDgEnF+GwkW/XfQgIxJHavbCeAjOANJnBNappYtsupkkgau7s9
			tZMMmTGL8Jxj775Smp62DAGJx5votVTPzJrHKiLwJpULqi5DqfzE5PgBBGO8o3DU
			aH8G6YzLqZgeNX1xToEKVLAT2hVZZQ4v/ZsjX5kbt1tqN541m5W1xbloi86Rku4J
			3O46KKwDEeERq2zmBcGtVcjmYFJ112ET7ounqp/9jtpqdaRX/OnGd7gjrydEoWKM
			VBJZDDDbsAOe27bRFDau
			-----END CERTIFICATE-----
			EOF
			]
			private_key = <<-EOF
			-----BEGIN PRIVATE KEY-----
			MIIEuwIBADANBgkqhkiG9w0BAQEFAASCBKUwggShAgEAAoIBAQCnUfnSg19ofGhn
			yYvG7I9Bt+lnaFbtmn7ZTjR6ASeqcFOTfqL79/D5R5en+j+dJa0FsoQfJ1yDjxHv
			ANwXBuhyGb8xLNdy3y9TlzP7BLGN9Zt2doJmSokuILIczpNxEnY0pSPXyNEJhH6x
			wDIH8I4nPaE76M1ZZ/n5pNxNNYkZUH+7UIk2m8LsNFtMpUmcelVTDN827DicncWv
			Jw46EqxiGTlLqnXzyb7pyh7BIcj5GyYyRfGBsOcc1geL+ci1e3XIkDi89kXe4qlo
			fSWx+2IwHYRFKLga6kQvHM7cd2XJnX23Pal6hWVSuvxaaeHatcxBcnsv5L+UvZUy
			KsTS7qBTAgMBAAECggEABnU+CLeEW7KNjw/y4qsnvlgcXJ7k2AfiBH4lvV3FC6mJ
			OESngsUnml9+hX+9q9GT84fX3KH2yqcfgJOOax8boqfGvt2ltSvTFk1cNsCQH9QO
			e4yIbO1MjSi65yy7+R3GzTJgh0gbdVwVTcQGylKpEe+phPfv0RcXyWBpFlvOHlly
			kHQJ8urM9KQ8Pi1blw9kSxXPr4cJsa9xRtATsupuuKGzXDtHiyhfS8dKfNjCG+05
			2pFBBoK94+8/2MG5O4zmFtQ+sy2mwZs4HRz5IaWQc6LLYHWTg8PUa8iwjY/o+tFB
			X9lIhMQ5tja3atuvdyQv6BzVHXESkCYThOqrhWTQmQKBgQDV7VTYU8ULbMukRAUg
			4FfafNxCbjIh02QeIUfkXKxA1E7bWr0Mx9nfe7mE1bem0x2FiLMNlbp/MMDO2Crv
			Ne2XIzgsOdqZIDYVDRBsLB5iJqlxW5Zir6qnsJZ/JX3I+KxzftTCDiy/Wa+jCR+7
			tke5Jlwz7aLe3aDYbntEL1QRqQKBgQDIOhxw4L0f5bVwzkUy3SEt0khrQptMzq0v
			S7+jAka6ls5Dom2cVma+Xvn0byjcMk3OBgOsvN/ywjhcri8Y7TKEcs72dQFERj3j
			hFqIk+k8b79eo6SnkEFDKyulw8gEenzv/11HXHud+9WyUgnMy45NxlqEw6FDhTQm
			RapL9HvXmwKBgQCirPMT/b+dTIIey8rKkU69Sq2DpqBgsIs1jkFJGl+yfL/qdjnE
			ekTneQI+TPZ22Ztda/IcpntHNR+pKyCa/vtJLvMMToI4ZxI5N9IBMBt6r8Ox+9+D
			8+ll0xbeYPgh11fsC8pmNrk4WU8CP3HuIFKyLMV4h4CO0SH68yixVPws4QKBgFEp
			7Tl8gG2bqg8OLlLN/JMceKqyF03tQZq4c/haBd3BH9+eyhvjkkZ9LYl+Pev0oEFx
			gq/U6Fr5i+tV2FWcYSv7dhXFnDvW1WOS1Tgj7RnImqR8ZVRfT3Uw3MKXOE9Ib7jB
			pUg2Hw4NdbSRONPBd+/jBfJncslyB4+0EbI0arcdAn9gDng5iZGdfaKEeNjdtsNZ
			9lr22zUDcejp08U3bIdQOGTOhs/FB2PSg4J7UpXMYaQMnSsQOBhf6Q1W0GRixIjO
			1gwyj6OOiknlVQrkE8xXq89BKRub035NMcRPP/QzqAaTNiSWwPHo1/VHY9qpxCY6
			k3bGtb64/xLdzdpD8rSi
			-----END PRIVATE KEY-----			
			EOF
		
			project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
		}
		`,
		projectName,
		newCertificateName,
	)
}
