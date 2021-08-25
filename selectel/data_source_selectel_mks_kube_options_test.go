package selectel

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
)

func TestAccMKSAvailableFeatureGatesV1Basic(t *testing.T) {
	var (
		project projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	kubeVersion := "1.16.3"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAvailableFeatureGatesV1BasicConfig(projectName, kubeVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("data.selectel_mks_available_feature_gates_v1.feature_gates_test", "available_feature_gates.0.kube_version_minor", kubeVersion),
					testFeatureGatesIsNotEmpty("data.selectel_mks_available_feature_gates_v1.feature_gates_test"),
				),
			},
		},
	})
}

func TestAccMKSAvailableFeatureGatesV1NoFilter(t *testing.T) {
	var (
		project projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAvailableFeatureGatesV1ConfigNoFilter(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testFeatureGatesNoFilter("data.selectel_mks_available_feature_gates_v1.feature_gates_test"),
				),
			},
		},
	})
}

func testFeatureGatesIsNotEmpty(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		availableFeatureGates, ok := rs.Primary.Attributes["available_feature_gates.0.names"]
		if !ok {
			return fmt.Errorf("attribute 'available_feature_gates' is not found")
		}
		if availableFeatureGates == "" {
			return fmt.Errorf("names is not set at 'available_feature_gates'")
		}

		return nil
	}
}
func testFeatureGatesNoFilter(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		availableFeatureGates, ok := rs.Primary.Attributes["available_feature_gates.0.names"]
		if !ok {
			return fmt.Errorf("attribute 'available_feature_gates' is not found")
		}
		if availableFeatureGates == "" {
			return fmt.Errorf("names is not set at 'available_feature_gates'")
		}

		fgLen, err := strconv.Atoi(rs.Primary.Attributes["available_feature_gates.#"])
		if err != nil {
			return fmt.Errorf("failed to get len of 'available_feature_gates': %w", err)
		}
		if fgLen <= 1 {
			return fmt.Errorf("received only one or less item in 'available_feature_gates'")
		}

		return nil
	}
}

func testAvailableFeatureGatesV1BasicConfig(projectName, kubeVersion string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

data "selectel_mks_available_feature_gates_v1" "feature_gates_test" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    kube_version_minor = "%s"
  }
}
`, projectName, kubeVersion)
}

func testAvailableFeatureGatesV1ConfigNoFilter(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

data "selectel_mks_available_feature_gates_v1" "feature_gates_test" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}
`, projectName)
}
