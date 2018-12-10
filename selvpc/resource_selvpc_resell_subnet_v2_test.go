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
	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/subnets"
	"github.com/stretchr/testify/assert"
)

func TestAccResellV2SubnetBasic(t *testing.T) {
	var subnet subnets.Subnet
	var project projects.Project
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

func TestResourceResellSubnetV2PrefixLengthFromCIDR(t *testing.T) {
	testingData := map[string]int{
		"192.0.2.100/29":   29,
		"192.0.2.200/28":   28,
		"203.0.113.10/24":  24,
		"203.0.113.129/25": 25,
	}

	for cidr, expected := range testingData {
		actual, err := resourceResellSubnetV2PrefixLengthFromCIDR(cidr)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}
}

func TestResourceResellSubnetV2GetIPVersionFromPrefixLength(t *testing.T) {
	testingData := map[int]string{
		29: string(selvpcclient.IPv4),
		28: string(selvpcclient.IPv4),
		48: string(selvpcclient.IPv6),
		64: string(selvpcclient.IPv6),
		24: string(selvpcclient.IPv4),
		25: string(selvpcclient.IPv4),
	}

	for prefixLength, expected := range testingData {
		actual := resourceResellSubnetV2GetIPVersionFromPrefixLength(prefixLength)

		assert.Equal(t, expected, actual)
	}
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
