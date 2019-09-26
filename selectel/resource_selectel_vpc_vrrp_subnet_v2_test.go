package selectel

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/vrrpsubnets"
)

func TestAccVPCV2VRRPSubnetBasic(t *testing.T) {
	var (
		vrrpSubnet vrrpsubnets.VRRPSubnet
		project    projects.Project
	)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCV2VRRPSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2VRRPSubnetBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckVPCV2VRRPSubnetExists("selectel_vpc_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", &vrrpSubnet),
					resource.TestCheckResourceAttr("selectel_vpc_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "master_region", "ru-1"),
					resource.TestCheckResourceAttr("selectel_vpc_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "slave_region", "ru-2"),
					resource.TestCheckResourceAttr("selectel_vpc_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "prefix_length", "29"),
					resource.TestCheckResourceAttr("selectel_vpc_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "ip_version", "ipv4"),
					resource.TestCheckResourceAttr("selectel_vpc_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "subnets.#", "2"),
					resource.TestCheckResourceAttr("selectel_vpc_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1", "status", "DOWN"),
				),
			},
		},
	})
}

func testAccCheckVPCV2VRRPSubnetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_vrrp_subnet_v2" {
			continue
		}

		_, _, err := vrrpsubnets.Get(ctx, resellV2Client, rs.Primary.ID)
		if err == nil {
			return errors.New("VRRP subnet still exists")
		}
	}

	return nil
}

func testAccCheckVPCV2VRRPSubnetExists(n string, vrrpSubnet *vrrpsubnets.VRRPSubnet) resource.TestCheckFunc {
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

func testAccVPCV2VRRPSubnetBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_vpc_vrrp_subnet_v2" "vrrp_subnet_tf_acc_test_1" {
  project_id    = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  master_region = "ru-1"
  slave_region  = "ru-2"
}`, projectName)
}
