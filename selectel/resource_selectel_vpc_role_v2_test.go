package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/roles"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/users"
)

func TestAccVPCV2RoleBasic(t *testing.T) {
	var (
		role    roles.Role
		project projects.Project
		user    users.User
	)
	projectName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")
	userPassword := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCV2RoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2RoleBasic(projectName, userName, userPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2RoleExists("selectel_vpc_role_v2.role_tf_acc_test_1", &role),
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckVPCV2UserExists("selectel_vpc_user_v2.user_tf_acc_test_1", &user),
					resource.TestCheckResourceAttrSet("selectel_vpc_role_v2.role_tf_acc_test_1", "project_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_role_v2.role_tf_acc_test_1", "user_id"),
				),
			},
		},
	})
}

func testAccCheckVPCV2RoleDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_role_v2" {
			continue
		}

		projectID, _, err := resourceVPCRoleV2ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}
		projectRoles, _, err := roles.ListProject(ctx, resellV2Client, projectID)
		if err == nil {
			if len(projectRoles) > 0 {
				return fmt.Errorf("there are still some roles in project '%s'", projectID)
			}
		}
	}

	return nil
}

func testAccCheckVPCV2RoleExists(n string, role *roles.Role) resource.TestCheckFunc {
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

		projectID, userID, err := resourceVPCRoleV2ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}
		projectRoles, _, err := roles.ListProject(ctx, resellV2Client, projectID)
		if err != nil {
			return errSearchingProjectRole(projectID, err)
		}

		found := false
		foundRoleIdx := 0
		for i, role := range projectRoles {
			if role.UserID == userID {
				found = true
				foundRoleIdx = i
			}
		}

		if !found {
			return errors.New("role not found")
		}

		*role = *projectRoles[foundRoleIdx]

		return nil
	}
}

func testAccVPCV2RoleBasic(projectName, userName, userPassword string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_vpc_user_v2" "user_tf_acc_test_1" {
  name        = "%s"
  password    = "%s"
}

resource "selectel_vpc_role_v2" "role_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  user_id    = "${selectel_vpc_user_v2.user_tf_acc_test_1.id}"
}`, projectName, userName, userPassword)
}
