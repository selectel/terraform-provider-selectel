package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSFirewallV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	nodeCount := 1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSFirewallV1Basic(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttrSet("selectel_dbaas_firewall_v1.firewall_tf_acc_test_1", "datastore_id"),
					resource.TestCheckResourceAttr("selectel_dbaas_firewall_v1.firewall_tf_acc_test_1", "ips.#", "2"),
				),
			},
		},
	})
}

func testAccDBaaSFirewallV1Basic(projectName, datastoreName string, nodeCount int) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  filter {
    engine = "postgresql"
    version = "16"
  }
}

resource "selectel_dbaas_datastore_v1" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor {
    vcpus = 2
    ram = 4096
    disk = 32
  }
}

resource "selectel_dbaas_firewall_v1" "firewall_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_datastore_v1.datastore_tf_acc_test_1.id}"
  ips = [ "127.0.0.1", "127.0.0.2" ]
}`, projectName, datastoreName, nodeCount)
}
