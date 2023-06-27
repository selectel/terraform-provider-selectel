package selectel

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v1 "github.com/selectel/craas-go/pkg"
	"github.com/selectel/craas-go/pkg/v1/registry"
	"github.com/selectel/go-selvpcclient/v2/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/v2/selvpcclient/resell/v2/tokens"
	"testing"
)

func TestAccCRaaSRegistryV1Basic(t *testing.T) {
	var (
		project       projects.Project
		craasRegistry registry.Registry
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	registryName := acctest.RandomWithPrefix("tf-acc-reg")

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

		var projectID string
		if id, ok := rs.Primary.Attributes["project_id"]; ok {
			projectID = id
		}

		config := testAccProvider.Meta().(*Config)
		resellV2Client := config.resellV2Client()
		ctx := context.Background()

		tokenOpts := tokens.TokenOpts{
			ProjectID: projectID,
		}
		token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
		if err != nil {
			return errCreatingObject(objectToken, err)
		}

		craasClient := v1.NewCRaaSClientV1(token.ID, craasV1Endpoint)
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
