package selectel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDedicatedOSV1Basic(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")
	osName := "Ubuntu"
	osVersion := "2204"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedOSV1Basic(projectID, osName, osVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccDedicatedOSV1Exists("data.selectel_dedicated_os_v1.os_tf_acc_test_1", osName, osVersion),
					resource.TestCheckResourceAttr("data.selectel_dedicated_os_v1.os_tf_acc_test_1", "os.0.name", osName),
					resource.TestCheckResourceAttr("data.selectel_dedicated_os_v1.os_tf_acc_test_1", "os.0.version_value", osVersion),
				),
			},
		},
	})
}

func testAccDedicatedOSV1Exists(
	n string, osName, osVersion string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dsClient := newTestDedicatedAPIClient(rs, testAccProvider)

		operatingSystemsFromAPI, _, err := dsClient.OperatingSystems(ctx, nil)
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

func testAccDedicatedOSV1Basic(projectID, osName, osVersion string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dedicated_os_v1" "os_tf_acc_test_1" {
  project_id = local.project_id

  filter {
    name             = "%s"
    version_value    = "%s"
  }
}
`, projectID, osName, osVersion)
}
