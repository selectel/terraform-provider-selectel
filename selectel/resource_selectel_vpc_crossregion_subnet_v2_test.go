package selectel

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
	"github.com/stretchr/testify/assert"
)

func TestAccVPCV2CrossRegionSubnetBasic(t *testing.T) {
	var (
		crossRegionSubnet crossregionsubnets.CrossRegionSubnet
		project           projects.Project
	)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCV2CrossRegionSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2CrossRegionSubnetBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckVPCV2CrossRegionSubnetExists("selectel_vpc_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", &crossRegionSubnet),
					resource.TestCheckResourceAttr("selectel_vpc_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", "cidr", "192.168.200.0/24"),
					resource.TestCheckResourceAttr("selectel_vpc_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", "regions.#", "2"),
					resource.TestCheckResourceAttr("selectel_vpc_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", "subnets.#", "2"),
					resource.TestCheckResourceAttr("selectel_vpc_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1", "status", "DOWN"),
				),
			},
		},
	})
}

func testAccCheckVPCV2CrossRegionSubnetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_crossregion_subnet_v2" {
			continue
		}

		_, _, err := crossregionsubnets.Get(ctx, resellV2Client, rs.Primary.ID)
		if err == nil {
			return errors.New("Cross-region subnet still exists")
		}
	}

	return nil
}

func testAccCheckVPCV2CrossRegionSubnetExists(n string, crossRegionSubnet *crossregionsubnets.CrossRegionSubnet) resource.TestCheckFunc {
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
			return errors.New("cross-region subnet not found")
		}

		*crossRegionSubnet = *foundCrossRegionSubnet

		return nil
	}
}

func testAccVPCV2CrossRegionSubnetBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_vpc_crossregion_subnet_v2" "crossregion_subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  cidr = "192.168.200.0/24"
  regions {
    region = "ru-1"
  }
  regions {
    region = "ru-3"
  }
}`, projectName)
}

func TestProjectIDFromSubnetsMaps(t *testing.T) {
	subnetsMaps := []map[string]interface{}{
		{
			"network_id":      "912bd5d0-cb11-4a7f-af7c-ea84c8e7db2e",
			"subnet_id":       "4912cca9-cad2-49c1-a69a-929cd4cf9559",
			"region":          "ru-2",
			"cidr":            "192.168.200.0/24",
			"vlan_id":         1003,
			"project_id":      "b63ab68796e34858befb8fa2a8b1e12a",
			"vtep_ip_address": "10.10.0.101",
		},
		{
			"network_id":      "954c6ebd-f923-4471-847a-e1be04af8952",
			"subnet_id":       "4754c984-bb91-4221-820c-ae2b0f64dae0",
			"region":          "ru-3",
			"cidr":            "192.168.200.0/24",
			"vlan_id":         1003,
			"project_id":      "b63ab68796e34858befb8fa2a8b1e12a",
			"vtep_ip_address": "10.10.0.201",
		},
	}

	expectedProjectID := "b63ab68796e34858befb8fa2a8b1e12a"

	actualProjectID, err := projectIDFromSubnetsMaps(subnetsMaps)

	assert.NoError(t, err)
	assert.Equal(t, expectedProjectID, actualProjectID)
}
