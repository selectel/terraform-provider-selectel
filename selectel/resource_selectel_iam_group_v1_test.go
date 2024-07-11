package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/iam-go/service/groups"
)

func TestAccIAMV1GroupBasic(t *testing.T) {
	var group groups.Group

	testName := "test-name"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1GroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1GroupBasic(testName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1GroupExists("selectel_iam_group_v1.group_tf_acc_test_1", &group),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_group_v1.group_tf_acc_test_1", "name", testName),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.0.scope"),
				),
			},
		},
	})
}

func TestAccIAMV1GroupUpdateRoles(t *testing.T) {
	var group groups.Group

	testName := "test-name"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1GroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1GroupBasic(testName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1GroupExists("selectel_iam_group_v1.group_tf_acc_test_1", &group),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_group_v1.group_tf_acc_test_1", "name", testName),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.0.scope"),
				),
			},
			{
				Config: testAccIAMV1GroupAssignRole(testName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1GroupExists("selectel_iam_group_v1.group_tf_acc_test_1", &group),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_group_v1.group_tf_acc_test_1", "name", testName),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.0.scope"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.1.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.1.scope"),
				),
			},
			{
				Config: testAccIAMV1GroupBasic(testName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMV1GroupExists("selectel_iam_group_v1.group_tf_acc_test_1", &group),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttr("selectel_iam_group_v1.group_tf_acc_test_1", "name", testName),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.0.role_name"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_v1.group_tf_acc_test_1", "role.0.scope"),
					resource.TestCheckNoResourceAttr("selectel_iam_group_v1.group_tf_acc_test_1", "role.1.role_name"),
					resource.TestCheckNoResourceAttr("selectel_iam_group_v1.group_tf_acc_test_1", "role.1.scope"),
				),
			},
		},
	})
}

func testAccCheckIAMV1GroupDestroy(s *terraform.State) error {
	iamClient, diagErr := getIAMClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get iamclient for test group object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_iam_group_v1" {
			continue
		}

		_, err := iamClient.Groups.Get(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("group still exists")
		}
	}

	return nil
}

func testAccCheckIAMV1GroupExists(n string, group *groups.Group) resource.TestCheckFunc {
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
			return fmt.Errorf("can't get iamclient for test group object")
		}

		g, err := iamClient.Groups.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return errors.New("group not found")
		}

		*group = g.Group

		return nil
	}
}

func testAccIAMV1GroupBasic(name string) string {
	return fmt.Sprintf(`
resource "selectel_iam_group_v1" "group_tf_acc_test_1" {
	name = "%s"
	role {
	  	role_name = "reader"
	  	scope = "account"
	}
}`, name)
}

func testAccIAMV1GroupAssignRole(name string) string {
	return fmt.Sprintf(`
	resource "selectel_iam_group_v1" "group_tf_acc_test_1" {
		name = "%s"
		role {
			role_name = "reader"
			scope = "account"
		}
		role {
			role_name = "billing"
			scope = "account"
		}
	}`, name)
}
