package selectel

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/licenses"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccVPCV2LicenseBasic(t *testing.T) {
	var (
		license licenses.License
		project projects.Project
	)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2LicenseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2LicenseBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckVPCV2LicenseExists("selectel_vpc_license_v2.license_tf_acc_test_1", &license),
					resource.TestCheckResourceAttr("selectel_vpc_license_v2.license_tf_acc_test_1", "region", "ru-1"),
					resource.TestCheckResourceAttr("selectel_vpc_license_v2.license_tf_acc_test_1", "type", "license_windows_2012_standard"),
					resource.TestCheckResourceAttr("selectel_vpc_license_v2.license_tf_acc_test_1", "status", "DOWN"),
					resource.TestCheckResourceAttr("selectel_vpc_license_v2.license_tf_acc_test_1", "port_id", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_license_v2.license_tf_acc_test_1", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_license_v2.license_tf_acc_test_1", "subnet_id"),
				),
			},
		},
	})
}

func testAccCheckVPCV2LicenseDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return fmt.Errorf("can't get selvpc client for test license object: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_license_v2" {
			continue
		}

		_, _, err := licenses.Get(selvpcClient, rs.Primary.ID)
		if err == nil {
			return errors.New("license still exists")
		}
	}

	return nil
}

func testAccCheckVPCV2LicenseExists(n string, license *licenses.License) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		selvpcClient, err := config.GetSelVPCClient()
		if err != nil {
			return fmt.Errorf("can't get selvpc client for test license object: %w", err)
		}

		foundLicense, _, err := licenses.Get(selvpcClient, rs.Primary.ID)
		if err != nil {
			return err
		}

		foundLicenseStrID := strconv.Itoa(foundLicense.ID)
		if foundLicenseStrID != rs.Primary.ID {
			return errors.New("license not found")
		}

		*license = *foundLicense

		return nil
	}
}

func testAccVPCV2LicenseBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

resource "selectel_vpc_license_v2" "license_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-1"
  type       = "license_windows_2012_standard"
}`, projectName)
}
