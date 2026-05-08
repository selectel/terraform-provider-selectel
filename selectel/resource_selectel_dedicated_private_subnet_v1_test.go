package selectel

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDedicatedPrivateSubnetV1Basic(t *testing.T) {
	var (
		project      projects.Project
		subnetID     string
		projectName  = acctest.RandomWithPrefix("tf-acc")
		locationName = "SPB-5"
		vlan         = "2989"
		subnet       = "10.10.10.0/24"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedPrivateSubnetV1Basic(projectName, locationName, vlan, subnet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDedicatedPrivateSubnetV1Exists("selectel_dedicated_private_subnet_v1.subnet_tf_acc_test_1", &subnetID),
					resource.TestCheckResourceAttrSet("selectel_dedicated_private_subnet_v1.subnet_tf_acc_test_1", "location_id"),
					resource.TestCheckResourceAttrSet("selectel_dedicated_private_subnet_v1.subnet_tf_acc_test_1", "vlan"),
					resource.TestCheckResourceAttrSet("selectel_dedicated_private_subnet_v1.subnet_tf_acc_test_1", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_dedicated_private_subnet_v1.subnet_tf_acc_test_1", "id"),
				),
			},
			{
				Config:   testAccDedicatedPrivateSubnetV1Basic(projectName, locationName, vlan, subnet),
				PlanOnly: true,
			},
			{
				Config:      testAccDedicatedPrivateSubnetV1InvalidCIDR(projectName, locationName, vlan),
				ExpectError: regexp.MustCompile("invalid CIDR format"),
			},
			{
				Config:      testAccDedicatedPrivateSubnetV1InvalidVLAN(projectName, locationName, subnet),
				ExpectError: regexp.MustCompile(`invalid parameters to filter on tag or vlan`),
			},
		},
	})
}

func testAccCheckDedicatedPrivateSubnetV1Exists(n string, subnetID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		*subnetID = rs.Primary.ID

		return nil
	}
}

func testAccDedicatedPrivateSubnetV1Basic(projectName, location, vlan, subnet string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

data "selectel_dedicated_location_v1" "location_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  filter {
    name = "%s"
  }
}

resource "selectel_dedicated_private_subnet_v1" "subnet_tf_acc_test_1" {
  location_id  = "${data.selectel_dedicated_location_v1.location_tf_acc_test_1.locations[0].id}"
  vlan         = "%s"
  subnet       = "%s"
}`, projectName, location, vlan, subnet)
}

func testAccDedicatedPrivateSubnetV1InvalidCIDR(projectName, location, vlan string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

data "selectel_dedicated_location_v1" "location_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  filter {
    name = "%s"
  }
}

resource "selectel_dedicated_private_subnet_v1" "subnet_tf_acc_test_1" {
  location_id  = "${data.selectel_dedicated_location_v1.location_tf_acc_test_1.locations[0].id}"
  vlan         = "%s"
  subnet       = "192.168.100.0/255"
}
`, projectName, location, vlan)
}

func testAccDedicatedPrivateSubnetV1InvalidVLAN(projectName, location, subnet string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

data "selectel_dedicated_location_v1" "location_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  filter {
    name = "%s"
  }
}

resource "selectel_dedicated_private_subnet_v1" "subnet_tf_acc_test_1" {
  location_id  = "${data.selectel_dedicated_location_v1.location_tf_acc_test_1.locations[0].id}"
  vlan         = "65000"
  subnet       = "%s"
}
`, projectName, location, subnet)
}
