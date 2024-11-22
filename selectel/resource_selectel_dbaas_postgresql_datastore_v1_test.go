package selectel

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSPostgreSQLDatastoreV1Basic(t *testing.T) {
	var (
		dbaasDatastore dbaas.Datastore
		project        projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	nodeCount := 1
	resizeNodeCount := 2

	updatedDatastoreName := acctest.RandomWithPrefix("tf-acc-ds-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSPostgreSQLDatastoreV1Basic(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSPostgreSQLDatastoreV1UpdateName(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSPostgreSQLDatastoreV1UpdatePooler(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "pooler.0.mode", "session"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "pooler.0.size", strconv.Itoa(50)),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSPostgreSQLDatastoreV1UpdateFirewall(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "pooler.0.mode", "session"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "pooler.0.size", strconv.Itoa(50)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSPostgreSQLDatastoreV1Resize(projectName, updatedDatastoreName, resizeNodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(resizeNodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(8192)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "pooler.0.mode", "session"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "pooler.0.size", strconv.Itoa(50)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSPostgreSQLDatastoreV1UpdateConfig(projectName, updatedDatastoreName, resizeNodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(resizeNodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(8192)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "pooler.0.mode", "session"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "pooler.0.size", strconv.Itoa(50)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(256)),
					resource.TestCheckResourceAttr("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(20)),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_postgresql_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
		},
	})
}

func testAccDBaaSPostgreSQLDatastoreV1Basic(projectName, datastoreName string, nodeCount int) string {
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
    version = "13"
  }
}

resource "selectel_dbaas_postgresql_datastore_v1" "datastore_tf_acc_test_1" {
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
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSPostgreSQLDatastoreV1UpdateName(projectName, datastoreName string, nodeCount int) string {
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
    version = "13"
  }
}

resource "selectel_dbaas_postgresql_datastore_v1" "datastore_tf_acc_test_1" {
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
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSPostgreSQLDatastoreV1UpdatePooler(projectName, datastoreName string, nodeCount int) string {
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
    version = "13"
  }
}

resource "selectel_dbaas_postgresql_datastore_v1" "datastore_tf_acc_test_1" {
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
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
  pooler {
    mode = "session"
    size = 50
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSPostgreSQLDatastoreV1UpdateFirewall(projectName, datastoreName string, nodeCount int) string {
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
    version = "13"
  }
}

resource "selectel_dbaas_postgresql_datastore_v1" "datastore_tf_acc_test_1" {
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
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
  pooler {
    mode = "session"
    size = 50
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSPostgreSQLDatastoreV1Resize(projectName, datastoreName string, nodeCount int) string {
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
    version = "13"
  }
}

resource "selectel_dbaas_postgresql_datastore_v1" "datastore_tf_acc_test_1" {
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
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
  pooler {
    mode = "session"
    size = 50
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSPostgreSQLDatastoreV1UpdateConfig(projectName, datastoreName string, nodeCount int) string {
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
    version = "13"
  }
}

resource "selectel_dbaas_postgresql_datastore_v1" "datastore_tf_acc_test_1" {
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
    xmloption = "content"
    work_mem = 256
    vacuum_cost_delay = 20
  }
  pooler {
    mode = "session"
    size = 50
  }
}`, projectName, datastoreName, nodeCount)
}
