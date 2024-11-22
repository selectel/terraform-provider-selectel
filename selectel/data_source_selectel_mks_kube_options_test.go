package selectel

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

const (
	dataSourceFeatureGates         = "selectel_mks_feature_gates_v1"
	dataSourceAdmissionControllers = "selectel_mks_admission_controllers_v1"
)

func TestAccMKSAvailableFeatureGatesV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	kubeVersion := testAccMKSClusterV1GetDefaultKubeVersion(t)
	kubeVersionMinor, err := kubeVersionTrimToMinor(kubeVersion)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testKubeOptionsV1BasicConfig(projectName, dataSourceFeatureGates, kubeVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr(getDataSourceName(dataSourceFeatureGates), "feature_gates.0.kube_version", kubeVersionMinor),
					testFeatureGatesIsNotEmpty(getDataSourceName(dataSourceFeatureGates)),
				),
			},
		},
	})
}

func TestAccMKSAvailableFeatureGatesV1NoFilter(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testKubeOptionsV1ConfigNoFilter(projectName, dataSourceFeatureGates),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testFeatureGatesNoFilter(getDataSourceName(dataSourceFeatureGates)),
				),
			},
		},
	})
}

func TestAccMKSAvailableAdmissionControllersV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	kubeVersion := testAccMKSClusterV1GetDefaultKubeVersion(t)
	kubeVersionMinor, err := kubeVersionTrimToMinor(kubeVersion)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testKubeOptionsV1BasicConfig(projectName, dataSourceAdmissionControllers, kubeVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr(getDataSourceName(dataSourceAdmissionControllers), "admission_controllers.0.kube_version", kubeVersionMinor),
					testAdmissionControllersIsNotEmpty(getDataSourceName(dataSourceAdmissionControllers)),
				),
			},
		},
	})
}

func TestAccMKSAvailableAdmissionControllersV1NoFilter(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testKubeOptionsV1ConfigNoFilter(projectName, dataSourceAdmissionControllers),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
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

func testKubeOptionsV1BasicConfig(projectName, dataSource, kubeVersion string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "%s" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    kube_version = "%s"
  }
}
`, projectName, dataSource, kubeVersion)
}

func testKubeOptionsV1ConfigNoFilter(projectName, dataSource string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "%s" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}
`, projectName, dataSource)
}

func getDataSourceName(dataSource string) string {
	return fmt.Sprintf("data.%s.dt", dataSource)
}
