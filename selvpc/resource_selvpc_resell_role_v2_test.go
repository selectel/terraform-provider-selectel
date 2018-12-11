package selvpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/roles"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/users"
)

func TestAccResellV2RoleBasic(t *testing.T) {
	var (
		role    roles.Role
		project projects.Project
		user    users.User
	)
	projectName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")
	userPassword := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2RoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2RoleBasic(projectName, userName, userPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2RoleExists("selvpc_resell_role_v2.role_tf_acc_test_1", &role),
					testAccCheckResellV2ProjectExists("selvpc_resell_project_v2.project_tf_acc_test_1", &project),
					testAccCheckResellV2UserExists("selvpc_resell_user_v2.user_tf_acc_test_1", &user),
					resource.TestCheckResourceAttrSet("selvpc_resell_role_v2.role_tf_acc_test_1", "project_id"),
					resource.TestCheckResourceAttrSet("selvpc_resell_role_v2.role_tf_acc_test_1", "user_id"),
				),
			},
		},
	})
}

func testAccCheckResellV2RoleDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selvpc_resell_role_v2" {
			continue
		}

		projectID, _, err := resourceResellRoleV2ParseID(rs.Primary.ID)
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

func testAccCheckResellV2RoleExists(n string, role *roles.Role) resource.TestCheckFunc {
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

		projectID, userID, err := resourceResellRoleV2ParseID(rs.Primary.ID)
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

func testAccResellV2RoleBasic(projectName, userName, userPassword string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selvpc_resell_user_v2" "user_tf_acc_test_1" {
  name        = "%s"
  password    = "%s"
}

resource "selvpc_resell_role_v2" "role_tf_acc_test_1" {
  project_id = "${selvpc_resell_project_v2.project_tf_acc_test_1.id}"
  user_id    = "${selvpc_resell_user_v2.user_tf_acc_test_1.id}"
}`, projectName, userName, userPassword)
}
