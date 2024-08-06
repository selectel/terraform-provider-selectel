package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/iam-go/service/serviceusers"
)

func TestAccIAMV1ServiceUserBasic(t *testing.T) {
	var serviceUser serviceusers.ServiceUser
	serviceUserName := acctest.RandomWithPrefix("tf-acc")
	serviceUserPassword := "A" + acctest.RandString(8) + "1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1ServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1ServiceUserBasic(serviceUserName, serviceUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1ServiceUserExists("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", &serviceUser),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "name", serviceUserName),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "password", serviceUserPassword),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.scope"),
				),
			},
		},
	})
}

func TestAccIAMV1ServiceUserUpdateRoles(t *testing.T) {
	var serviceUser serviceusers.ServiceUser
	serviceUserName := acctest.RandomWithPrefix("tf-acc")
	serviceUserPassword := "A" + acctest.RandString(8) + "1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1ServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1ServiceUserBasic(serviceUserName, serviceUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1ServiceUserExists("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", &serviceUser),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "name", serviceUserName),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "password", serviceUserPassword),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.scope"),
				),
			},
			{
				Config: testAccIAMV1ServiceUserAssignRole(serviceUserName, serviceUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1ServiceUserExists("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", &serviceUser),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "name", serviceUserName),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "password", serviceUserPassword),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.scope"),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.1.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.1.scope"),
				),
			},
			{
				Config: testAccIAMV1ServiceUserBasic(serviceUserName, serviceUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1ServiceUserExists("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", &serviceUser),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "name", serviceUserName),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "password", serviceUserPassword),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.scope"),
					resource.TestCheckNoResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.1.role_name"),
					resource.TestCheckNoResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.1.scope"),
				),
			},
		},
	})
}

func TestAccIAMV1ServiceUserUpdateName(t *testing.T) {
	var serviceUser serviceusers.ServiceUser
	serviceUserName := acctest.RandomWithPrefix("tf-acc")
	serviceUserPassword := "A" + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1ServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1ServiceUserBasic(serviceUserName, serviceUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1ServiceUserExists("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", &serviceUser),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "name", serviceUserName),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "password", serviceUserPassword),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.scope"),
				),
			},
			{
				Config: testAccIAMV1ServiceUserUpdateName(serviceUserPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1ServiceUserExists("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", &serviceUser),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "name", "NewName"),
					resource.TestCheckResourceAttr("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "password", serviceUserPassword),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1", "role.0.scope"),
				),
			},
		},
	})
}

func testAccCheckIAMV1ServiceUserDestroy(s *terraform.State) error {
	iamClient, diagErr := getIAMClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get iamclient for test serviceUser object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_iam_serviceuser_v1" {
			continue
		}

		_, err := iamClient.ServiceUsers.Get(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("serviceUser still exists")
		}
	}

	return nil
}

func testAccCheckIAMV1ServiceUserExists(n string, serviceUser *serviceusers.ServiceUser) resource.TestCheckFunc {
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
			return fmt.Errorf("can't get iamclient for test serviceUser object")
		}

		su, err := iamClient.ServiceUsers.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return errors.New("serviceUser not found")
		}

		*serviceUser = su.ServiceUser

		return nil
	}
}

func testAccIAMV1ServiceUserBasic(userName, userPassword string) string {
	return fmt.Sprintf(`
resource "selectel_iam_serviceuser_v1" "serviceuser_tf_acc_test_1" {
  name        = "%s"
  password    = "%s"
  role {
    role_name = "reader"
    scope = "account"
  }
}`, userName, userPassword)
}

func testAccIAMV1ServiceUserAssignRole(userName, userPassword string) string {
	return fmt.Sprintf(`
resource "selectel_iam_serviceuser_v1" "serviceuser_tf_acc_test_1" {
  name        = "%s"
  password    = "%s"
  role {
    role_name = "reader"
    scope = "account"
  }
  role {
    role_name = "iam_admin"
    scope = "account"
  }
}`, userName, userPassword)
}

func testAccIAMV1ServiceUserUpdateName(userPassword string) string {
	return fmt.Sprintf(`
resource "selectel_iam_serviceuser_v1" "serviceuser_tf_acc_test_1" {
  name        = "NewName"
  password    = "%s"
  role {
    role_name = "reader"
    scope = "account"
  }
}`, userPassword)
}
