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

func TestAccServersOSV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	osName := "Ubuntu"
	osVersion := "2204"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServersOSV1Basic(projectName, osName, osVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccServersOSV1Exists("data.selectel_servers_os_v1.os_tf_acc_test_1", osName, osVersion),
					resource.TestCheckResourceAttr("data.selectel_servers_os_v1.os_tf_acc_test_1", "os.0.name", osName),
					resource.TestCheckResourceAttr("data.selectel_servers_os_v1.os_tf_acc_test_1", "os.0.version", osVersion),
				),
			},
		},
	})
}

func testAccServersOSV1Exists(
	n string, osName, osVersion string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dsClient := newTestServersAPIClient(rs, testAccProvider)

		operatingSystemsFromAPI, _, err := dsClient.OperatingSystems(ctx)
		if err != nil {
			return err
		}

		osFromAPI := operatingSystemsFromAPI.FindOneByNameAndVersion(osName, osVersion)

		if osFromAPI == nil {
			return fmt.Errorf("os %s %s not found", osName, osVersion)
		}

		return nil
	}
}

func testAccServersOSV1Basic(projectName, osName, osVersion string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

data "selectel_servers_os_v1" "os_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"

  filter {
    name             = "%s"
    version          = "%s"
  }
}
`, projectName, osName, osVersion)
}
