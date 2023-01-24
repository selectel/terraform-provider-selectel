package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMKSKubeVersionsV1DataSourceBasic(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMKSKubeVersionsV1Basic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMKSKubeVersionsV1("data.selectel_mks_kube_versions_v1.kube_versions_tf_acc_test_1"),
				),
			},
		},
	})
}

func testAccCheckMKSKubeVersionsV1(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find kube versions data source: %s", name)
		}
		if _, ok := rs.Primary.Attributes["latest_version"]; !ok {
			return errors.New("empty 'latest_version' field in kube versions data source")
		}
		if _, ok := rs.Primary.Attributes["default_version"]; !ok {
			return errors.New("empty 'default_version' field in kube versions data source")
		}
		if _, ok := rs.Primary.Attributes["versions.#"]; !ok {
			return errors.New("empty 'versions' field in kube versions data source")
		}

		return nil
	}
}

func testAccMKSKubeVersionsV1Basic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_mks_kube_versions_v1" "kube_versions_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}
`, projectName)
}
