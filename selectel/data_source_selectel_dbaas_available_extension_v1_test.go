package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSAvailableExtensionsV1Basic(t *testing.T) {
	var (
		dbaasAvailableExtensions []dbaas.AvailableExtension
		project                  projects.Project
	)

	const availableExtensionName = "hstore"

	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSAvailableExtensionsV1Basic(projectName, availableExtensionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
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

func testAccDBaaSAvailableExtensionsV1Basic(projectName, name string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dbaas_available_extension_v1" "available_extension_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    name = "%s"
  }
}
`, projectName, name)
}
