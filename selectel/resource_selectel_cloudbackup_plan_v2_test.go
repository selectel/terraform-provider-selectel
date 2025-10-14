package selectel

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccCloudBackupPlanV2(t *testing.T) {
	var (
		project projects.Project

		projectName = acctest.RandomWithPrefix("tf-acc")

		name                                        = "tf-backup-plan"
		backupMode                                  = "full"
		description, descriptionUpdated             = "Weekly full at 04:00 on Sunday", "Daily full at 06:30"
		fullBackupsAmount, fullBackupsAmountUpdated = 4, 2
		scheduleType                                = "crontab"
		schedulePattern, schedulePatternUpdated     = "0 4 * * 0", "30 4 * * 0"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			// create case
			{
				Config: testAccCloudBackupPlanV2(projectName, name, backupMode, description, scheduleType, schedulePattern, fullBackupsAmount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.name", name),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.backup_mode", backupMode),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.description", description),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.full_backups_amount", strconv.Itoa(fullBackupsAmount)),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.schedule_type", scheduleType),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.schedule_pattern", schedulePattern),
				),
			},
			// update cases
			{
				Config: testAccCloudBackupPlanV2(projectName, name, backupMode, descriptionUpdated, scheduleType, schedulePatternUpdated, fullBackupsAmountUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.name", name),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.backup_mode", backupMode),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.description", descriptionUpdated),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.full_backups_amount", strconv.Itoa(fullBackupsAmountUpdated)),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.schedule_type", scheduleType),
					resource.TestCheckResourceAttr("data.selectel_cloudbackup_plan_v2.plans", "plans.0.list.0.schedule_pattern", schedulePatternUpdated),
				),
			},
		},
	})
}

func testAccCloudBackupPlanV2(
	projectName, name, backupMode, description, scheduleType, schedulePattern string, maxBackups int,
) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
	name        = "%s"
}

provider "openstack" {
  tenant_id   = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
}

resource "openstack_blockstorage_volume_v3" "volume_1" {
  region = "ru-1"
  name   = "volume_1"
  size   = 3
}

resource "selectel_cloudbackup_plan_v2" "backupplan_1" {
	project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
    region = "ru-1"

	name = "%s"
	backup_mode = "%s"
	description = "%s"
	full_backups_amount = %d
	schedule_type = "%s"
	schedule_pattern = "%s"
	resources {
		resource {
      		type = "OS::Cinder::Volume"
      		id   = openstack_blockstorage_volume_v3.volume_1.id
      		name = openstack_blockstorage_volume_v3.volume_1.name
    	}
	}
}

data "selectel_cloudbackup_plan_v2" "plans" {
	project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
    region = "ru-1"

	filter {
		name = "%s"
		volume_name = openstack_blockstorage_volume_v3.volume_1.name
	}

	depends_on = [selectel_cloudbackup_plan_v2.backupplan_1]
}
`, projectName, name, backupMode, description, maxBackups, scheduleType, schedulePattern, name)
}
