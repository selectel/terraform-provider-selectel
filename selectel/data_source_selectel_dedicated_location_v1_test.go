package selectel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDedicatedLocationV1Basic(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")
	locationName := "MSK-2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedLocationV1Basic(projectID, locationName),
				Check: resource.ComposeTestCheckFunc(
					testAccDedicatedLocationV1Exists("data.selectel_dedicated_location_v1.location_tf_acc_test_1", locationName),
					resource.TestCheckResourceAttr("data.selectel_dedicated_location_v1.location_tf_acc_test_1", "locations.0.name", locationName),
				),
			},
		},
	})
}

func testAccDedicatedLocationV1Exists(
	n string, locationName string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dsClient := newTestDedicatedAPIClient(rs, testAccProvider)

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

func testAccDedicatedLocationV1Basic(projectID, locationName string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dedicated_location_v1" "location_tf_acc_test_1" {
  project_id = local.project_id
  filter {
    name = "%s"
  }
}
`, projectID, locationName)
}
