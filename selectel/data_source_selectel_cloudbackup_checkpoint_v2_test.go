package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccCloudBackupCheckpointV2Basic(t *testing.T) {
	var (
		project     projects.Project
		projectName = acctest.RandomWithPrefix("tf-acc")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudBackupCheckpointV2Basic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
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

func testAccCloudBackupCheckpointV2Basic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_cloudbackup_checkpoint_v2" "checkpoints" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-1"

  filter {
     plan_name = "non-existing-plan-name"
     volume_name = "non-existing-volume-name"
  }
}
`, projectName)
}
