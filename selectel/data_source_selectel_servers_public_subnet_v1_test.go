package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccServersPublicSubnetV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	locationName := "MSK-2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServersPublicSubnetV1Basic(projectName, locationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("data.selectel_servers_public_subnet_v1.public_subnet_tf_acc_test_1", "subnets.#", "0"),
				),
			},
		},
	})
}

func testAccServersPublicSubnetV1Basic(projectName, locationName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

data "selectel_servers_location_v1" "location_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  filter {
    name = "%s"
  }
}

data "selectel_servers_public_subnet_v1" "public_subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"

  filter {
    location_id = "${data.selectel_servers_location_v1.location_tf_acc_test_1.locations.0.id}"
  }
}
`, projectName, locationName)
}
