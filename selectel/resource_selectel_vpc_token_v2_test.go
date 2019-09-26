package selectel

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
)

func TestAccVPCV2TokenBasic(t *testing.T) {
	var project projects.Project
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccSelectelPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2TokenBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttrSet("selectel_vpc_token_v2.token_tf_acc_test_1", "project_id"),
				),
			},
		},
	})
}

func TestAccVPCV2TokenAccount(t *testing.T) {
	accountName := ""

	if selToken := os.Getenv("SEL_TOKEN"); strings.ContainsAny(selToken, "_") {
		accountName = strings.Split(selToken, "_")[1]
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccSelectelPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2TokenAccount(accountName),
				Check:  resource.TestCheckResourceAttrSet("selectel_vpc_token_v2.token_tf_acc_test_1", "account_name"),
			},
		},
	})
}

func testAccVPCV2TokenBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_vpc_token_v2" "token_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
}
`, projectName)
}

func testAccVPCV2TokenAccount(accountName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_token_v2" "token_tf_acc_test_1" {
  account_name = "%s"
}
`, accountName)
}
