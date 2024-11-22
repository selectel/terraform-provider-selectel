package selectel

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/subnets"
)

func TestAccVPCV2SubnetBasic(t *testing.T) {
	var (
		subnet  subnets.Subnet
		project projects.Project
	)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2SubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2SubnetBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckVPCV2SubnetExists("selectel_vpc_subnet_v2.subnet_tf_acc_test_1", &subnet),
					resource.TestCheckResourceAttr("selectel_vpc_subnet_v2.subnet_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_vpc_subnet_v2.subnet_tf_acc_test_1", "status", "DOWN"),
				),
			},
		},
	})
}

func testAccCheckVPCV2SubnetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return fmt.Errorf("can't get selvpc client for test subnet object: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_resell_subnet_v2" {
			continue
		}

		_, _, err := subnets.Get(selvpcClient, rs.Primary.ID)
		if err == nil {
			return errors.New("subnet still exists")
		}
	}

	return nil
}

func testAccCheckVPCV2SubnetExists(n string, subnet *subnets.Subnet) resource.TestCheckFunc {
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
			return fmt.Errorf("can't get selvpc client for test subnet object: %w", err)
		}

		foundSubnet, _, err := subnets.Get(selvpcClient, rs.Primary.ID)
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

func testAccVPCV2SubnetBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}`, projectName)
}
