package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/floatingips"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccVPCV2FloatingIPBasic(t *testing.T) {
	var (
		floatingip floatingips.FloatingIP
		project    projects.Project
	)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2FloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2FloatingIPBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckVPCV2FloatingIPExists("selectel_vpc_floatingip_v2.floatingip_tf_acc_test_1", &floatingip),
					resource.TestCheckResourceAttr("selectel_vpc_floatingip_v2.floatingip_tf_acc_test_1", "region", "ru-2"),
					resource.TestCheckResourceAttr("selectel_vpc_floatingip_v2.floatingip_tf_acc_test_1", "status", "DOWN"),
				),
			},
		},
	})
}

func testAccCheckVPCV2FloatingIPDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return fmt.Errorf("can't get selvpc client for test floatingip object: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_floatingip_v2" {
			continue
		}

		_, _, err := floatingips.Get(selvpcClient, rs.Primary.ID)
		if err == nil {
			return errors.New("floatingip still exists")
		}
	}

	return nil
}

func testAccCheckVPCV2FloatingIPExists(n string, floatingip *floatingips.FloatingIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		selvpcClient, err := config.GetSelVPCClient()
		if err != nil {
			return fmt.Errorf("can't get selvpc client for test floatingip object: %w", err)
		}

		foundFloatingIP, _, err := floatingips.Get(selvpcClient, rs.Primary.ID)
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

func testAccVPCV2FloatingIPBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

resource "selectel_vpc_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-2"
}`, projectName)
}
