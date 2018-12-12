package selvpc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
)

func TestAccResellV2TokenBasic(t *testing.T) {
	var project projects.Project
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccSelVPCPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2TokenBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2ProjectExists("selvpc_resell_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttrSet("selvpc_resell_token_v2.token_tf_acc_test_1", "project_id"),
				),
			},
		},
	})
}

func TestAccResellV2TokenAccount(t *testing.T) {
	accountName := "79414"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccSelVPCPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2TokenAccount(accountName),
				Check:  resource.TestCheckResourceAttrSet("selvpc_resell_token_v2.token_tf_acc_test_1", "account_name"),
			},
		},
	})
}

func testAccResellV2TokenBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selvpc_resell_token_v2" "token_tf_acc_test_1" {
  project_id = "${selvpc_resell_project_v2.project_tf_acc_test_1.id}"
}
`, projectName)
}

func testAccResellV2TokenAccount(accountName string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_token_v2" "token_tf_acc_test_1" {
  account_name = "%s"
}
`, accountName)
}
