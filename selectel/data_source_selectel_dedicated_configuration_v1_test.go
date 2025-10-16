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

func TestAccDedicatedConfigurationV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	configurationName := "EL50-SSD"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedConfigurationV1Basic(projectName, configurationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccDedicatedConfigurationV1Exists("data.selectel_dedicated_configuration_v1.server_configuration_tf_acc_test_1", configurationName),
					resource.TestCheckResourceAttr("data.selectel_dedicated_configuration_v1.server_configuration_tf_acc_test_1", "configurations.0.name", configurationName),
				),
			},
		},
	})
}

func testAccDedicatedConfigurationV1Exists(
	n string, serverName string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dsClient := newTestDedicatedAPIClient(rs, testAccProvider)

		serversFromAPI, _, err := dsClient.ServersRaw(ctx)
		if err != nil {
			return err
		}

		var srvFromAPI map[string]interface{}
		for _, srv := range serversFromAPI {
			name, _ := srv["name"].(string)
			if name == serverName {
				srvFromAPI = srv
			}
		}

		if srvFromAPI == nil {
			return fmt.Errorf("server %s not found", serverName)
		}

		return nil
	}
}

func testAccDedicatedConfigurationV1Basic(projectName, configurationName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dedicated_configuration_v1" "server_configuration_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"

  deep_filter = "{\"name\": \"%s\"}"
}
`, projectName, configurationName)
}
