package selectel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
)

func TestAccDBaaSAvailableExtensionsV1Basic(t *testing.T) {
	var dbaasAvailableExtensions []dbaas.AvailableExtension

	const availableExtensionName = "hstore"

	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSAvailableExtensionsV1Basic(projectID, availableExtensionName),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSAvailableExtensionsV1Exists("data.selectel_dbaas_available_extension_v1.available_extension_tf_acc_test_1", &dbaasAvailableExtensions),
					resource.TestCheckResourceAttr("data.selectel_dbaas_available_extension_v1.available_extension_tf_acc_test_1", "available_extensions.0.name", availableExtensionName),
				),
			},
		},
	})
}

func testAccDBaaSAvailableExtensionsV1Exists(n string, dbaasAvailableExtensions *[]dbaas.AvailableExtension) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dbaasClient, err := newTestDBaaSClient(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		availableExtensions, err := dbaasClient.AvailableExtensions(ctx)
		if err != nil {
			return err
		}

		*dbaasAvailableExtensions = availableExtensions

		return nil
	}
}

func testAccDBaaSAvailableExtensionsV1Basic(projectID, name string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dbaas_available_extension_v1" "available_extension_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-3"
  filter {
    name = "%s"
  }
}
`, projectID, name)
}
