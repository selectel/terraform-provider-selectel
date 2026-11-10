package selectel

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudBackupCheckpointV2Basic(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudBackupCheckpointV2Basic(projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.selectel_cloudbackup_checkpoint_v2.checkpoints", "checkpoints.list.#", "0",
					),
					resource.TestCheckResourceAttr(
						"data.selectel_cloudbackup_checkpoint_v2.checkpoints", "checkpoints.total.#", "0",
					),
				),
			},
		},
	})
}

func testAccCloudBackupCheckpointV2Basic(projectID string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_cloudbackup_checkpoint_v2" "checkpoints" {
  project_id = local.project_id
  region = "ru-1"

  filter {
     plan_name = "non-existing-plan-name"
     volume_name = "non-existing-volume-name"
  }
}
`, projectID)
}
