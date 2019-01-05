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
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/vrrpsubnets"
)

func TestAccResellV2VRRPSubnetBasic(t *testing.T) {
	var (
		vrrpSubnet vrrpsubnets.VRRPSubnet
		project    projects.Project
	)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2VRRPSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2VRRPSubnetBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2ProjectExists("selvpc_resell_project_v2.project_tf_acc_test_1", &project),
					testAccCheckResellV2VRRPSubnetExists("selvpc_resell_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", &vrrpSubnet),
					resource.TestCheckResourceAttr("selvpc_resell_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "master_region", "ru-1"),
					resource.TestCheckResourceAttr("selvpc_resell_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "slave_region", "ru-2"),
					resource.TestCheckResourceAttr("selvpc_resell_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "prefix_length", "29"),
					resource.TestCheckResourceAttr("selvpc_resell_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "ip_version", "ipv4"),
					resource.TestCheckResourceAttr("selvpc_resell_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "subnets.#", "2"),
					resource.TestCheckResourceAttr("selvpc_resell_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "status", "DOWN"),
				),
			},
		},
	})
}

func testAccCheckResellV2VRRPSubnetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selvpc_resell_vrrp_subnet_v2" {
			continue
		}

		_, _, err := vrrpsubnets.Get(ctx, resellV2Client, rs.Primary.ID)
		if err == nil {
			return errors.New("VRRP subnet still exists")
		}
	}

	return nil
}

func testAccCheckResellV2VRRPSubnetExists(n string, vrrpSubnet *vrrpsubnets.VRRPSubnet) resource.TestCheckFunc {
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

		foundVRRPSubnet, _, err := vrrpsubnets.Get(ctx, resellV2Client, rs.Primary.ID)
		if err != nil {
			return err
		}

		if strconv.Itoa(foundVRRPSubnet.ID) != rs.Primary.ID {
			return errors.New("VRRP subnet not found")
		}

		*vrrpSubnet = *foundVRRPSubnet

		return nil
	}
}

func testAccResellV2VRRPSubnetBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selvpc_resell_vrrp_subnet_v2" "vrrp_subnet_tf_acc_test_1" {
  project_id    = "${selvpc_resell_project_v2.project_tf_acc_test_1.id}"
  master_region = "ru-1"
  slave_region  = "ru-2"
}`, projectName)
}
