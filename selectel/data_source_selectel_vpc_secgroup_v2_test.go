package selectel

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccVPCSecGroupV2DataSourceName = "data.selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1"

func TestAccVPCSecGroupV2DataSource(t *testing.T) {
	projectID, region := vpcV2TestEnv()
	secGroupName := acctest.RandomWithPrefix("tf-acc-secgroup")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecGroupV2DataSource(projectID, region, secGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						testAccVPCSecGroupV2DataSourceName, "secgroup_id",
						testAccVPCSecGroupV2ResourceName, "id",
					),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2DataSourceName, "name", secGroupName),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2DataSourceName, "description", "created by the acceptance test"),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2DataSourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2DataSourceName, "all_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(testAccVPCSecGroupV2DataSourceName, "all_tags.*", secGroupName),
					resource.TestCheckResourceAttrPair(
						testAccVPCSecGroupV2DataSourceName, "shared",
						testAccVPCSecGroupV2ResourceName, "shared",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCSecGroupV2DataSourceName, "security_group_rules.#",
						testAccVPCSecGroupV2ResourceName, "security_group_rules.#",
					),
				),
			},
			{
				Config:      testAccVPCSecGroupV2DataSourceAmbiguous(projectID, region, secGroupName),
				ExpectError: regexp.MustCompile(`2 VPC security groups found`),
			},
			{
				Config:      testAccVPCSecGroupV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no VPC security group found`),
			},
		},
	})
}

func testAccVPCSecGroupV2DataSource(projectID, region, name string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%[1]s"
  name        = "%[2]s"
  description = "created by the acceptance test"
  stateful    = true
  tags        = ["tf-acc", "%[2]s"]
}

data "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%[1]s"
  tags        = ["%[2]s"]
  description = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.description
}
`, region, name)
}

func testAccVPCSecGroupV2DataSourceAmbiguous(projectID, region, name string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "%[2]s"
  tags       = ["%[2]s"]
}

resource "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "%[2]s-twin"
  tags       = ["%[2]s"]
}

data "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  tags       = ["%[2]s"]

  depends_on = [
    selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1,
    selectel_vpc_secgroup_v2.secgroup_tf_acc_test_2,
  ]
}
`, region, name)
}

func testAccVPCSecGroupV2DataSourceNoMatch(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
data "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-secgroup"
}
`, region)
}
