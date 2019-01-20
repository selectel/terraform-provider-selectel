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
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/crossregionsubnets"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
)

func TestAccResellV2CrossRegionSubnetBasic(t *testing.T) {
	var (
		crossRegionSubnet crossregionsubnets.CrossRegionSubnet
		project           projects.Project
	)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2CrossRegionSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2CrossRegionSubnetBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2ProjectExists("selvpc_resell_project_v2.project_tf_acc_test_1", &project),
					testAccCheckResellV2CrossRegionSubnetExists("selvpc_resell_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", &crossRegionSubnet),
					resource.TestCheckResourceAttr("selvpc_resell_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", "cidr", "192.168.200.0/24"),
					resource.TestCheckResourceAttr("selvpc_resell_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", "regions.#", "2"),
					resource.TestCheckResourceAttr("selvpc_resell_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", "subnets.#", "2"),
					resource.TestCheckResourceAttr("selvpc_resell_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", "status", "DOWN"),
				),
			},
		},
	})
}

func testAccCheckResellV2CrossRegionSubnetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selvpc_resell_crossregion_subnet_v2" {
			continue
		}

		_, _, err := crossregionsubnets.Get(ctx, resellV2Client, rs.Primary.ID)
		if err == nil {
			return errors.New("Cross-region subnet still exists")
		}
	}

	return nil
}

func testAccCheckResellV2CrossRegionSubnetExists(n string, crossRegionSubnet *crossregionsubnets.CrossRegionSubnet) resource.TestCheckFunc {
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

		foundCrossRegionSubnet, _, err := crossregionsubnets.Get(ctx, resellV2Client, rs.Primary.ID)
		if err != nil {
			return err
		}

		if strconv.Itoa(foundCrossRegionSubnet.ID) != rs.Primary.ID {
			return errors.New("Cross-region subnet not found")
		}

		*crossRegionSubnet = *foundCrossRegionSubnet

		return nil
	}
}

func testAccResellV2CrossRegionSubnetBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selvpc_resell_crossregion_subnet_v2" "crossregion_subnet_tf_acc_test_1" {
  project_id    = "${selvpc_resell_project_v2.project_tf_acc_test_1.id}"
  regions = [
    {
      region = "ru-1"
    },
    {
      region = "ru-3"
    },
  ]
  cidr = "192.168.200.0/24"
}`, projectName)
}
