package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/iam-go/service/ec2"
)

func TestAccIAMV1EC2Basic(t *testing.T) {
	var ec2 ec2.Credential
	ec2Name := acctest.RandomWithPrefix("tf-acc")
	projectName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")
	userPassword := "A" + acctest.RandString(8) + "1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1EC2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1EC2Basic(projectName, userName, userPassword, ec2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1EC2Exists("selectel_iam_ec2_v1.ec2_tf_acc_test_1", &ec2),
					resource.TestCheckResourceAttrSet("selectel_iam_ec2_v1.ec2_tf_acc_test_1", "user_id"),
					resource.TestCheckResourceAttrSet("selectel_iam_ec2_v1.ec2_tf_acc_test_1", "project_id"),
					resource.TestCheckResourceAttrSet("selectel_iam_ec2_v1.ec2_tf_acc_test_1", "access_key"),
					resource.TestCheckResourceAttrSet("selectel_iam_ec2_v1.ec2_tf_acc_test_1", "secret_key"),
					resource.TestCheckResourceAttr("selectel_iam_ec2_v1.ec2_tf_acc_test_1", "name", ec2Name),
				),
			},
		},
	})
}

func testAccCheckIAMV1EC2Destroy(s *terraform.State) error {
	iamClient, diagErr := getIAMClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get iamclient for test ec2 object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_iam_ec2_v1" {
			continue
		}

		credentialsList, _ := iamClient.EC2.List(context.Background(), rs.Primary.Attributes["user_id"])
		for _, cred := range credentialsList {
			if cred.AccessKey == rs.Primary.ID {
				return errors.New("ec2 still exists")
			}
		}
	}

	return nil
}

func testAccCheckIAMV1EC2Exists(n string, ec2Credential *ec2.Credential) resource.TestCheckFunc {
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
			return fmt.Errorf("can't get  iamclient for test ec2 object")
		}

		credentialsList, _ := iamClient.EC2.List(context.Background(), rs.Primary.Attributes["user_id"])
		var neededEC2 ec2.Credential
		for _, cred := range credentialsList {
			if cred.Name == rs.Primary.Attributes["name"] {
				neededEC2 = cred
				break
			}
		}

		if neededEC2.AccessKey != rs.Primary.ID {
			return errors.New("ec2 not found")
		}

		*ec2Credential = neededEC2

		return nil
	}
}

func testAccIAMV1EC2Basic(projectName, userName, userPassword, ec2Name string) string {
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

resource "selectel_iam_ec2_v1" "ec2_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  user_id    = "${selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1.id}"
  name       = "%s"
}`, projectName, userName, userPassword, ec2Name)
}
