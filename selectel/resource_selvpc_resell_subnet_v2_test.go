package selvpc

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/subnets"
)

func TestAccResellV2SubnetBasic(t *testing.T) {
	var (
		subnet  subnets.Subnet
		project projects.Project
	)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2SubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2SubnetBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2ProjectExists("selvpc_resell_project_v2.project_tf_acc_test_1", &project),
					testAccCheckResellV2SubnetExists("selvpc_resell_subnet_v2.subnet_tf_acc_test_1", &subnet),
					resource.TestCheckResourceAttr("selvpc_resell_subnet_v2.subnet_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selvpc_resell_subnet_v2.subnet_tf_acc_test_1", "status", "DOWN"),
				),
			},
		},
	})
}

func testAccCheckResellV2SubnetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selvpc_resell_subnet_v2" {
			continue
		}

		_, _, err := subnets.Get(ctx, resellV2Client, rs.Primary.ID)
		if err == nil {
			return errors.New("subnet still exists")
		}
	}

	return nil
}

func testAccCheckResellV2SubnetExists(n string, subnet *subnets.Subnet) resource.TestCheckFunc {
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

		foundSubnet, _, err := subnets.Get(ctx, resellV2Client, rs.Primary.ID)
		if err != nil {
			return err
		}

		if strconv.Itoa(foundSubnet.ID) != rs.Primary.ID {
			return errors.New("subnet not found")
		}

		*subnet = *foundSubnet

		return nil
	}
}

func testAccResellV2SubnetBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selvpc_resell_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selvpc_resell_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}`, projectName)
}
