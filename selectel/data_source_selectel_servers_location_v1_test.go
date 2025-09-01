package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccServersLocationV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	locationName := "MSK-2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServersLocationV1Basic(projectName, locationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccServersLocationV1Exists("data.selectel_servers_location_v1.location_tf_acc_test_1", locationName),
					resource.TestCheckResourceAttr("data.selectel_servers_location_v1.location_tf_acc_test_1", "locations.0.name", locationName),
				),
			},
		},
	})
}

func testAccServersLocationV1Exists(
	n string, locationName string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dsClient := newTestServersAPIClient(rs, testAccProvider)

		locationsFromAPI, _, err := dsClient.Locations(ctx)
		if err != nil {
			return err
		}

		locFromAPI := locationsFromAPI.FindOneByName(locationName)

		if locFromAPI == nil {
			return fmt.Errorf("location %s not found", locationName)
		}

		return nil
	}
}

func testAccServersLocationV1Basic(projectName, locationName string) string {
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
`, projectName, locationName)
}
