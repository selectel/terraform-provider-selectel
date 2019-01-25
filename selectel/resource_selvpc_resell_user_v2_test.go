package selvpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/users"
)

func TestAccResellV2UserBasic(t *testing.T) {
	var user users.User
	userName := acctest.RandomWithPrefix("tf-acc")
	userNameUpdated := acctest.RandomWithPrefix("tf-acc")
	userPassword := acctest.RandString(8)
	userPasswordUpdated := acctest.RandString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2UserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2UserBasic(userName, userPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2UserExists("selvpc_resell_user_v2.user_tf_acc_test_1", &user),
					resource.TestCheckResourceAttr("selvpc_resell_user_v2.user_tf_acc_test_1", "name", userName),
					resource.TestCheckResourceAttr("selvpc_resell_user_v2.user_tf_acc_test_1", "password", userPassword),
					resource.TestCheckResourceAttr("selvpc_resell_user_v2.user_tf_acc_test_1", "enabled", "true"),
				),
			},
			{
				Config: testAccResellV2UserBasic(userNameUpdated, userPassword),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selvpc_resell_user_v2.user_tf_acc_test_1", "name", userNameUpdated),
					resource.TestCheckResourceAttr(
						"selvpc_resell_user_v2.user_tf_acc_test_1", "password", userPassword),
					resource.TestCheckResourceAttr(
						"selvpc_resell_user_v2.user_tf_acc_test_1", "enabled", "true"),
				),
			},
			{
				Config: testAccResellV2UserBasic(userNameUpdated, userPasswordUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selvpc_resell_user_v2.user_tf_acc_test_1", "name", userNameUpdated),
					resource.TestCheckResourceAttr(
						"selvpc_resell_user_v2.user_tf_acc_test_1", "password", userPasswordUpdated),
					resource.TestCheckResourceAttr(
						"selvpc_resell_user_v2.user_tf_acc_test_1", "enabled", "true"),
				),
			},
			{
				Config: testAccResellV2UserDisabled(userNameUpdated, userPasswordUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selvpc_resell_user_v2.user_tf_acc_test_1", "name", userNameUpdated),
					resource.TestCheckResourceAttr(
						"selvpc_resell_user_v2.user_tf_acc_test_1", "password", userPasswordUpdated),
					resource.TestCheckResourceAttr(
						"selvpc_resell_user_v2.user_tf_acc_test_1", "enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckResellV2UserDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selvpc_resell_user_v2" {
			continue
		}

		userList, _, err := users.List(ctx, resellV2Client)
		if err != nil {
			return err
		}

		found := false
		for _, user := range userList {
			if user.ID == rs.Primary.ID {
				found = true
			}
		}

		if found {
			return errors.New("user still exists")
		}
	}

	return nil
}

func testAccCheckResellV2UserExists(n string, user *users.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		resellV2Client := config.resellV2Client()
		ctx := context.Background()

		userList, _, err := users.List(ctx, resellV2Client)
		if err != nil {
			return err
		}

		found := false
		foundUserIdx := 0
		for i, resellV2User := range userList {
			if resellV2User.ID == rs.Primary.ID {
				found = true
				foundUserIdx = i
			}
		}

		if !found {
			return errors.New("user not found")
		}

		*user = *userList[foundUserIdx]

		return nil
	}
}

func testAccResellV2UserBasic(userName, userPassword string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_user_v2" "user_tf_acc_test_1" {
  name        = "%s"
  password    = "%s"
}`, userName, userPassword)
}

func testAccResellV2UserDisabled(userName, userPassword string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_user_v2" "user_tf_acc_test_1" {
  name        = "%s"
  password    = "%s"
  enabled     = false
}`, userName, userPassword)
}
