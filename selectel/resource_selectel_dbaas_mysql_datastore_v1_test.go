package selectel

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSMySQLDatastoreV1Basic(t *testing.T) {
	var (
		dbaasDatastore dbaas.Datastore
		project        projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	nodeCount := 1
	resizeNodeCount := 3

	updatedDatastoreName := acctest.RandomWithPrefix("tf-acc-ds-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSMySQLDatastoreV1Basic(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.innodb_checksum_algorithm", "strict_innodb"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.auto_increment_offset", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.autocommit", "false"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSMySQLDatastoreV1UpdateName(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.innodb_checksum_algorithm", "strict_innodb"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.auto_increment_offset", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.autocommit", "false"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSMySQLDatastoreV1UpdateFirewall(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.innodb_checksum_algorithm", "strict_innodb"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.auto_increment_offset", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.autocommit", "false"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSMySQLDatastoreV1Resize(projectName, updatedDatastoreName, resizeNodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(resizeNodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(8192)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.innodb_checksum_algorithm", "strict_innodb"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.auto_increment_offset", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.autocommit", "false"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSMySQLDatastoreV1UpdateConfig(projectName, updatedDatastoreName, resizeNodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(resizeNodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(8192)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.innodb_checksum_algorithm", "strict_innodb"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.auto_increment_offset", strconv.Itoa(4)),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "config.autocommit", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
		},
	})
}

func testAccDBaaSMySQLDatastoreV1Basic(projectName, datastoreName string, nodeCount int) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  filter {
    engine = "mysql"
    version = "8"
  }
}

resource "selectel_dbaas_mysql_datastore_v1" "datastore_tf_acc_test_1" {
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
  config = {
    innodb_checksum_algorithm = "strict_innodb"
	auto_increment_offset = 2
	autocommit = false
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSMySQLDatastoreV1UpdateName(projectName, datastoreName string, nodeCount int) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  filter {
    engine = "mysql"
    version = "8"
  }
}

resource "selectel_dbaas_mysql_datastore_v1" "datastore_tf_acc_test_1" {
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
  config = {
    innodb_checksum_algorithm = "strict_innodb"
	auto_increment_offset = 2
	autocommit = false
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSMySQLDatastoreV1UpdateFirewall(projectName, datastoreName string, nodeCount int) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  filter {
    engine = "mysql"
    version = "8"
  }
}

resource "selectel_dbaas_mysql_datastore_v1" "datastore_tf_acc_test_1" {
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
  config = {
    innodb_checksum_algorithm = "strict_innodb"
	auto_increment_offset = 2
	autocommit = false
  }
  firewall {
    ips = [ "127.0.0.1", "127.0.0.2" ]
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSMySQLDatastoreV1Resize(projectName, datastoreName string, nodeCount int) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  filter {
    engine = "mysql"
    version = "8"
  }
}

resource "selectel_dbaas_mysql_datastore_v1" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor {
    vcpus = 2
    ram = 8192
    disk = 32
  }
  config = {
    innodb_checksum_algorithm = "strict_innodb"
	auto_increment_offset = 2
	autocommit = false
  }
  firewall {
    ips = [ "127.0.0.1", "127.0.0.2" ]
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSMySQLDatastoreV1UpdateConfig(projectName, datastoreName string, nodeCount int) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  filter {
    engine = "mysql"
    version = "8"
  }
}

resource "selectel_dbaas_mysql_datastore_v1" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor {
    vcpus = 2
    ram = 8192
    disk = 32
  }
  config = {
    innodb_checksum_algorithm = "strict_innodb"
	auto_increment_offset = 4
	autocommit = true
  }
  firewall {
    ips = [ "127.0.0.1", "127.0.0.2" ]
  }
}`, projectName, datastoreName, nodeCount)
}
