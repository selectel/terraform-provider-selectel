package selvpc

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/licenses"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
)

func TestAccResellV2LicenseBasic(t *testing.T) {
	var license licenses.License
	var project projects.Project
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2LicenseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2LicenseBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2ProjectExists("selvpc_resell_project_v2.project_tf_acc_test_1", &project),
					testAccCheckResellV2LicenseExists("selvpc_resell_license_v2.license_tf_acc_test_1", &license),
					resource.TestCheckResourceAttr("selvpc_resell_license_v2.license_tf_acc_test_1", "region", "ru-1"),
					resource.TestCheckResourceAttr("selvpc_resell_license_v2.license_tf_acc_test_1", "type", "license_windows_2012_standard"),
					resource.TestCheckResourceAttr("selvpc_resell_license_v2.license_tf_acc_test_1", "status", "DOWN"),
				),
			},
		},
	})
}

func testAccCheckResellV2LicenseDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selvpc_resell_license_v2" {
			continue
		}

		_, _, err := licenses.Get(ctx, resellV2Client, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("license still exists")
		}
	}

	return nil
}

func testAccCheckResellV2LicenseExists(n string, license *licenses.License) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		resellV2Client := config.resellV2Client()
		ctx := context.Background()

		foundLicense, _, err := licenses.Get(ctx, resellV2Client, rs.Primary.ID)
		if err != nil {
			return err
		}

		foundLicenseStrID := strconv.Itoa(foundLicense.ID)
		if foundLicenseStrID != rs.Primary.ID {
			return fmt.Errorf("license not found")
		}

		*license = *foundLicense

		return nil
	}
}

func testAccResellV2LicenseBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

resource "selvpc_resell_license_v2" "license_tf_acc_test_1" {
  project_id = "${selvpc_resell_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-1"
  type       = "license_windows_2012_standard"
}`, projectName)
}
