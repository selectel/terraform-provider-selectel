package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/craas-go/pkg/v1/registry"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

const craasV1RegistryHostName = "https://cr.selcloud.ru"

func TestAccCRaaSRegistryV1Basic(t *testing.T) {
	var (
		project       projects.Project
		craasRegistry registry.Registry
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	registryName := acctest.RandomWithPrefix("tf-acc-reg")
	registryEndpoint := fmt.Sprintf("%s/%s", craasV1RegistryHostName, registryName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCRaaSRegistryV1Basic(projectName, registryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckCRaaSRegistryV1Exists("selectel_craas_registry_v1.registry_tf_acc_test_1", &craasRegistry),
					resource.TestCheckResourceAttr("selectel_craas_registry_v1.registry_tf_acc_test_1", "name", registryName),
					resource.TestCheckResourceAttr("selectel_craas_registry_v1.registry_tf_acc_test_1", "status", "ACTIVE"),
					resource.TestCheckResourceAttr("selectel_craas_registry_v1.registry_tf_acc_test_1", "endpoint", registryEndpoint),
				),
			},
		},
	})
}

func testAccCheckCRaaSRegistryV1Exists(n string, craasRegistry *registry.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		ctx := context.Background()

		craasClient, err := newCRaaSTestClient(rs, testAccProvider)
		if err != nil {
			return err
		}

		foundRegistry, _, err := registry.Get(ctx, craasClient, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundRegistry.ID != rs.Primary.ID {
			return errors.New("registry not found")
		}

		*craasRegistry = *foundRegistry

		return nil
	}
}

func testAccCRaaSRegistryV1Basic(projectName, registryName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

resource "selectel_craas_registry_v1" "registry_tf_acc_test_1" {
  name       = "%s"
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
}

`, projectName, registryName)
}
