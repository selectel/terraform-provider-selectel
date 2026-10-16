package selectel

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudBackupPlanV2Basic(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudBackupPlanV2Basic(projectID),
				Check: resource.ComposeTestCheckFunc(
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

func testAccCloudBackupPlanV2Basic(projectID string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_cloudbackup_plan_v2" "plans" {
  project_id = local.project_id
  region     = "ru-1"

  filter {
    name = "non-existent"
    volume_name = "non-existent-volume"
    status = "started"
  }
}
`, projectID)
}
