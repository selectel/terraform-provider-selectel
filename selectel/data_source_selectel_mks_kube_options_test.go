package selectel

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	dataSourceFeatureGates         = "selectel_mks_feature_gates_v1"
	dataSourceAdmissionControllers = "selectel_mks_admission_controllers_v1"
)

func TestAccMKSAvailableFeatureGatesV1Basic(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")
	kubeVersion := testAccMKSClusterV1GetDefaultKubeVersion(t)
	kubeVersionMinor, err := kubeVersionTrimToMinor(kubeVersion)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKubeOptionsV1BasicConfig(projectID, dataSourceFeatureGates, kubeVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(getDataSourceName(dataSourceFeatureGates), "feature_gates.0.kube_version", kubeVersionMinor),
					testFeatureGatesIsNotEmpty(getDataSourceName(dataSourceFeatureGates)),
				),
			},
		},
	})
}

func TestAccMKSAvailableFeatureGatesV1NoFilter(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKubeOptionsV1ConfigNoFilter(projectID, dataSourceFeatureGates),
				Check: resource.ComposeTestCheckFunc(
					testFeatureGatesNoFilter(getDataSourceName(dataSourceFeatureGates)),
				),
			},
		},
	})
}

func TestAccMKSAvailableAdmissionControllersV1Basic(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")
	kubeVersion := testAccMKSClusterV1GetDefaultKubeVersion(t)
	kubeVersionMinor, err := kubeVersionTrimToMinor(kubeVersion)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKubeOptionsV1BasicConfig(projectID, dataSourceAdmissionControllers, kubeVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(getDataSourceName(dataSourceAdmissionControllers), "admission_controllers.0.kube_version", kubeVersionMinor),
					testAdmissionControllersIsNotEmpty(getDataSourceName(dataSourceAdmissionControllers)),
				),
			},
		},
	})
}

func TestAccMKSAvailableAdmissionControllersV1NoFilter(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKubeOptionsV1ConfigNoFilter(projectID, dataSourceAdmissionControllers),
				Check: resource.ComposeTestCheckFunc(
					testAdmissionControllersNoFilter(getDataSourceName(dataSourceAdmissionControllers)),
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

		availableFeatureGates, ok := rs.Primary.Attributes["feature_gates.#"]
		if !ok {
			return fmt.Errorf("attribute 'feature_gates' is not found")
		}
		if availableFeatureGates == "" {
			return fmt.Errorf("names is not set at 'feature_gates'")
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

		availableFeatureGates, ok := rs.Primary.Attributes["feature_gates.0.names.#"]
		if !ok {
			return fmt.Errorf("attribute 'feature_gates' is not found")
		}
		if availableFeatureGates == "" {
			return fmt.Errorf("names is not set at 'feature_gates'")
		}

		fgCount, err := strconv.Atoi(availableFeatureGates)
		if err != nil {
			return fmt.Errorf("failed to get len of 'feature_gates': %w", err)
		}
		if fgCount <= 1 {
			return fmt.Errorf("received only one or less item in 'feature_gates'")
		}

		return nil
	}
}

func testAdmissionControllersIsNotEmpty(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		availableAdmissionControllers, ok := rs.Primary.Attributes["admission_controllers.#"]
		if !ok {
			return fmt.Errorf("attribute 'admission_controllers' is not found")
		}
		if availableAdmissionControllers == "" {
			return fmt.Errorf("names is not set at 'admission_controllers'")
		}

		return nil
	}
}

func testAdmissionControllersNoFilter(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		availableAdmissionControllers, ok := rs.Primary.Attributes["admission_controllers.0.names.#"]
		if !ok {
			return fmt.Errorf("attribute 'admission_controllers' is not found")
		}
		if availableAdmissionControllers == "" {
			return fmt.Errorf("names is not set at 'admission_controllers'")
		}

		acCount, err := strconv.Atoi(availableAdmissionControllers)
		if err != nil {
			return fmt.Errorf("failed to get len of 'admission_controllers': %w", err)
		}
		if acCount <= 1 {
			return fmt.Errorf("received only one or less item in 'admission_controllers'")
		}

		return nil
	}
}

func testKubeOptionsV1BasicConfig(projectID, dataSource, kubeVersion string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "%s" "dt" {
  project_id = local.project_id
  region     = "ru-3"
  filter {
    kube_version = "%s"
  }
}
`, projectID, dataSource, kubeVersion)
}

func testKubeOptionsV1ConfigNoFilter(projectID, dataSource string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "%s" "dt" {
  project_id = local.project_id
  region     = "ru-3"
}
`, projectID, dataSource)
}

func getDataSourceName(dataSource string) string {
	return fmt.Sprintf("data.%s.dt", dataSource)
}
