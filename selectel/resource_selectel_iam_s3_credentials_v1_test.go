package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/iam-go/service/s3credentials"
)

func TestAccIAMV1S3CredentialsBasic(t *testing.T) {
	var s3credential s3credentials.Credential
	s3CredsName := acctest.RandomWithPrefix("tf-acc")
	projectName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")
	userPassword := "A" + acctest.RandString(8) + "1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1S3CredentialsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1S3CredentialsBasic(projectName, userName, userPassword, s3CredsName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1S3CredentialsExists("selectel_iam_s3_credentials_v1.s3_creds_tf_acc_test_1", &s3credential),
					resource.TestCheckResourceAttrSet("selectel_iam_s3_credentials_v1.s3_creds_tf_acc_test_1", "user_id"),
					resource.TestCheckResourceAttrSet("selectel_iam_s3_credentials_v1.s3_creds_tf_acc_test_1", "project_id"),
					resource.TestCheckResourceAttrSet("selectel_iam_s3_credentials_v1.s3_creds_tf_acc_test_1", "secret_key"),
					resource.TestCheckResourceAttrSet("selectel_iam_s3_credentials_v1.s3_creds_tf_acc_test_1", "access_key"),
					resource.TestCheckResourceAttr("selectel_iam_s3_credentials_v1.s3_creds_tf_acc_test_1", "name", s3CredsName),
				),
			},
		},
	})
}

func testAccCheckIAMV1S3CredentialsDestroy(s *terraform.State) error {
	iamClient, diagErr := getIAMClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get iamclient for test s3 object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_iam_s3_credentials_v1" {
			continue
		}

		response, _ := iamClient.S3Credentials.List(context.Background(), rs.Primary.Attributes["user_id"])
		for _, cred := range response.Credentials {
			if cred.AccessKey == rs.Primary.ID {
				return errors.New("s3 credentials still exist")
			}
		}
	}

	return nil
}

func testAccCheckIAMV1S3CredentialsExists(n string, s3Credential *s3credentials.Credential) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		iamClient, diagErr := getIAMClient(testAccProvider.Meta())
		if diagErr != nil {
			return fmt.Errorf("can't get iamclient for test s3 credentials object")
		}

		response, _ := iamClient.S3Credentials.List(context.Background(), rs.Primary.Attributes["user_id"])
		var neededS3Credentials s3credentials.Credential
		for _, cred := range response.Credentials {
			if cred.Name == rs.Primary.Attributes["name"] {
				neededS3Credentials = cred
				break
			}
		}

		if neededS3Credentials.AccessKey != rs.Primary.ID {
			return errors.New("s3 credentials not found")
		}

		*s3Credential = neededS3Credentials

		return nil
	}
}

func testAccIAMV1S3CredentialsBasic(projectName, userName, userPassword, s3CredentialsName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

resource "selectel_iam_serviceuser_v1" "serviceuser_tf_acc_test_1" {
  name        = "%s"
  password    = "%s"
  role {
    role_name = "member"
    scope = "account"
  }
}

resource "selectel_iam_s3_credentials_v1" "s3_creds_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  user_id    = "${selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1.id}"
  name       = "%s"
}`, projectName, userName, userPassword, s3CredentialsName)
}
