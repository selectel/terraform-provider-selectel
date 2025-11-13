package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccCloudBackupPlanV2Basic(t *testing.T) {
	var (
		project     projects.Project
		projectName = acctest.RandomWithPrefix("tf-acc")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudBackupPlanV2Basic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr(
						"data.selectel_cloudbackup_plan_v2.plans", "plans.list.#", "0",
					),
					resource.TestCheckResourceAttr(
						"data.selectel_cloudbackup_plan_v2.plans", "plans.total.#", "0",
					),
				),
			},
		},
	})
}

func testAccCloudBackupPlanV2Basic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_cloudbackup_plan_v2" "plans" {
  project_id = "53c3fcf719044d1487e6a9d72d66b0a8"
  region     = "ru-1"

  filter {
    name = "non-existent"
    volume_name = "non-existent-volume"
    status = "started"
  }
}
`, projectName)
}
