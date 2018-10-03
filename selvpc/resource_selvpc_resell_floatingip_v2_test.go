package selvpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/floatingips"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
)

func TestAccResellV2FloatingIPBasic(t *testing.T) {
	var floatingip floatingips.FloatingIP
	var project projects.Project
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2FloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2FloatingIPBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2ProjectExists("selvpc_resell_project_v2.project_tf_acc_test_1", &project),
					testAccCheckResellV2FloatingIPExists("selvpc_resell_floatingip_v2.floatingip_tf_acc_test_1", &floatingip),
					resource.TestCheckResourceAttr("selvpc_resell_floatingip_v2.floatingip_tf_acc_test_1", "region", "ru-2"),
					resource.TestCheckResourceAttr("selvpc_resell_floatingip_v2.floatingip_tf_acc_test_1", "status", "DOWN"),
				),
			},
		},
	})
}

func testAccCheckResellV2FloatingIPDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selvpc_resell_floatingip_v2" {
			continue
		}

		_, _, err := floatingips.Get(ctx, resellV2Client, rs.Primary.ID)
		if err == nil {
			return errors.New("floatingip still exists")
		}
	}

	return nil
}

func testAccCheckResellV2FloatingIPExists(n string, floatingip *floatingips.FloatingIP) resource.TestCheckFunc {
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

		foundFloatingIP, _, err := floatingips.Get(ctx, resellV2Client, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundFloatingIP.ID != rs.Primary.ID {
			return errors.New("floatingip not found")
		}

		*floatingip = *foundFloatingIP

		return nil
	}
}

func testAccResellV2FloatingIPBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

resource "selvpc_resell_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id = "${selvpc_resell_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-2"
}`, projectName)
}
