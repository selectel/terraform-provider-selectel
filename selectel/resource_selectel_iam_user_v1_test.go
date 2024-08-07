package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/iam-go/service/users"
)

func TestAccIAMV1UserBasic(t *testing.T) {
	var user users.User
	userEmail := acctest.RandomWithPrefix("tf-acc") + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1UserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1UserBasic(userEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1UserExists("selectel_iam_user_v1.user_tf_acc_test_1", &user),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_user_v1.user_tf_acc_test_1", "email", userEmail),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.0.scope"),
				),
			},
		},
	})
}

func TestAccIAMV1UserUpdateRoles(t *testing.T) {
	var user users.User
	userEmail := acctest.RandomWithPrefix("tf-acc") + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1UserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1UserBasic(userEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1UserExists("selectel_iam_user_v1.user_tf_acc_test_1", &user),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_user_v1.user_tf_acc_test_1", "email", userEmail),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.0.scope"),
				),
			},
			{
				Config: testAccIAMV1UserAssignRole(userEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1UserExists("selectel_iam_user_v1.user_tf_acc_test_1", &user),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_user_v1.user_tf_acc_test_1", "email", userEmail),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.0.scope"),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.1.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.1.scope"),
				),
			},
			{
				Config: testAccIAMV1UserBasic(userEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1UserExists("selectel_iam_user_v1.user_tf_acc_test_1", &user),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_user_v1.user_tf_acc_test_1", "email", userEmail),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_user_v1.user_tf_acc_test_1", "role.0.scope"),
					resource.TestCheckNoResourceAttr("selectel_iam_user_v1.user_tf_acc_test_1", "role.1.role_name"),
					resource.TestCheckNoResourceAttr("selectel_iam_user_v1.user_tf_acc_test_1", "role.1.scope"),
				),
			},
		},
	})
}

func testAccCheckIAMV1UserDestroy(s *terraform.State) error {
	iamClient, diagErr := getIAMClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get iamclient for test user object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_iam_user_v1" {
			continue
		}

		_, err := iamClient.Users.Get(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("user still exists")
		}
	}

	return nil
}

func testAccCheckIAMV1UserExists(n string, user *users.User) resource.TestCheckFunc {
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
			return fmt.Errorf("can't get iamclient for test user object")
		}

		u, err := iamClient.Users.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return errors.New("user not found")
		}

		*user = u.User

		return nil
	}
}

func testAccIAMV1UserBasic(userEmail string) string {
	return fmt.Sprintf(`
resource "selectel_iam_user_v1" "user_tf_acc_test_1" {
	email = "%s"
	role {
	  	role_name = "reader"
	  	scope = "account"
	}
}`, userEmail)
}

func testAccIAMV1UserAssignRole(userEmail string) string {
	return fmt.Sprintf(`
	resource "selectel_iam_user_v1" "user_tf_acc_test_1" {
		email = "%s"
		role {
			role_name = "reader"
			scope = "account"
		}
		role {
			role_name = "billing"
			scope = "account"
		}
	}`, userEmail)
}
